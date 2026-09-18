"""Hidden test: ``get_parameter_source()`` must be correct while type
conversion and callbacks run, and feature-switch groups must keep the winning
option's source after parsing."""

import os

import click
from click.core import ParameterSource
from click.testing import CliRunner


class Recording(click.ParamType):
    name = "recording"

    def convert(self, value, param, ctx):
        src = ctx.get_parameter_source(param.name)
        return f"{value}|{src.name if src else 'None'}"


def _cli_with_recording_type():
    @click.command()
    @click.option("--a", type=Recording(), default="dflt")
    @click.option("--b", type=Recording(), envvar="HIDDEN_B")
    @click.option("--c", type=Recording())
    def cli(a, b, c):
        click.echo(f"a={a}")
        click.echo(f"b={b}")
        click.echo(f"c={c}")

    return cli


def test_convert_sees_default_source():
    result = CliRunner().invoke(_cli_with_recording_type(), [])
    assert result.exit_code == 0, result.output
    assert "a=dflt|DEFAULT" in result.output
    assert "b=None" in result.output
    assert "c=None" in result.output


def test_convert_sees_commandline_source():
    result = CliRunner().invoke(
        _cli_with_recording_type(), ["--a", "x", "--b", "y", "--c", "z"]
    )
    assert result.exit_code == 0, result.output
    assert "a=x|COMMANDLINE" in result.output
    assert "b=y|COMMANDLINE" in result.output
    assert "c=z|COMMANDLINE" in result.output


def test_convert_sees_environment_source():
    result = CliRunner().invoke(
        _cli_with_recording_type(), [], env={"HIDDEN_B": "from-env"}
    )
    assert result.exit_code == 0, result.output
    assert "b=from-env|ENVIRONMENT" in result.output


def test_convert_sees_default_map_source():
    result = CliRunner().invoke(
        _cli_with_recording_type(), [], default_map={"c": "mapped"}
    )
    assert result.exit_code == 0, result.output
    assert "c=mapped|DEFAULT_MAP" in result.output


def test_eager_callback_sees_source():
    seen = {}

    def cb(ctx, param, value):
        src = ctx.get_parameter_source(param.name)
        seen["callback"] = src.name if src else None
        return value

    @click.command()
    @click.option("--verbose/--quiet", default=False, is_eager=True, callback=cb)
    @click.pass_context
    def cli(ctx, verbose):
        src = ctx.get_parameter_source("verbose")
        click.echo(f"final={src.name}")

    result = CliRunner().invoke(cli, [])
    assert result.exit_code == 0, result.output
    assert seen["callback"] == "DEFAULT"
    assert "final=DEFAULT" in result.output

    result = CliRunner().invoke(cli, ["--verbose"])
    assert result.exit_code == 0, result.output
    assert seen["callback"] == "COMMANDLINE"
    assert "final=COMMANDLINE" in result.output


def test_non_eager_callback_sees_source():
    seen = {}

    def cb(ctx, param, value):
        src = ctx.get_parameter_source(param.name)
        seen["callback"] = src.name if src else None
        return value

    @click.command()
    @click.option("--level", default=3, callback=cb)
    def cli(level):
        click.echo(str(level))

    result = CliRunner().invoke(cli, ["--level", "5"])
    assert result.exit_code == 0, result.output
    assert seen["callback"] == "COMMANDLINE"

    result = CliRunner().invoke(cli, [])
    assert result.exit_code == 0, result.output
    assert seen["callback"] == "DEFAULT"


def test_debug_guard_pattern_respects_environment(monkeypatch):
    monkeypatch.delenv("HIDDEN_APP_DEBUG", raising=False)

    def set_debug(ctx, param, value):
        src = ctx.get_parameter_source(param.name)
        if src in (ParameterSource.DEFAULT, ParameterSource.DEFAULT_MAP):
            return None
        os.environ["HIDDEN_APP_DEBUG"] = "1" if value else "0"
        return value

    @click.command()
    @click.option(
        "--debug/--no-debug",
        default=False,
        is_eager=True,
        expose_value=False,
        callback=set_debug,
    )
    def cli():
        click.echo(f"DEBUG={os.environ.get('HIDDEN_APP_DEBUG', '')}")

    monkeypatch.setenv("HIDDEN_APP_DEBUG", "1")
    result = CliRunner().invoke(cli, [])
    assert result.exit_code == 0, result.output
    assert result.output.strip() == "DEBUG=1"

    result = CliRunner().invoke(cli, ["--no-debug"])
    assert result.exit_code == 0, result.output
    assert result.output.strip() == "DEBUG=0"

    result = CliRunner().invoke(cli, ["--debug"])
    assert result.exit_code == 0, result.output
    assert result.output.strip() == "DEBUG=1"


def _feature_switch_cli():
    @click.command()
    @click.option("--without-xyz", "enable_xyz", flag_value=False)
    @click.option("--with-xyz", "enable_xyz", flag_value=True, default=True)
    @click.pass_context
    def cli(ctx, enable_xyz):
        src = ctx.get_parameter_source("enable_xyz")
        click.echo(f"value={enable_xyz!r} source={src.name}")

    return cli


def test_feature_switch_default_wins():
    result = CliRunner().invoke(_feature_switch_cli(), [])
    assert result.exit_code == 0, result.output
    assert result.output.strip() == "value=True source=DEFAULT"


def test_feature_switch_commandline_wins():
    result = CliRunner().invoke(_feature_switch_cli(), ["--without-xyz"])
    assert result.exit_code == 0, result.output
    assert result.output.strip() == "value=False source=COMMANDLINE"


def test_feature_switch_loser_from_default_map_does_not_clobber_source():
    result = CliRunner().invoke(
        _feature_switch_cli(), ["--without-xyz"], default_map={"enable_xyz": True}
    )
    assert result.exit_code == 0, result.output
    assert result.output.strip() == "value=False source=COMMANDLINE"


def test_feature_switch_loser_from_envvar_does_not_clobber_source():
    @click.command()
    @click.option("--without-xyz", "enable_xyz", flag_value=False)
    @click.option(
        "--with-xyz", "enable_xyz", flag_value=True, default=True, envvar="HIDDEN_XYZ"
    )
    @click.pass_context
    def cli(ctx, enable_xyz):
        src = ctx.get_parameter_source("enable_xyz")
        click.echo(f"value={enable_xyz!r} source={src.name}")

    result = CliRunner().invoke(cli, ["--without-xyz"], env={"HIDDEN_XYZ": "1"})
    assert result.exit_code == 0, result.output
    assert result.output.strip() == "value=False source=COMMANDLINE"


def test_feature_switch_env_winner_keeps_environment_source():
    @click.command()
    @click.option(
        "--with-xyz", "enable_xyz", flag_value=True, default=True, envvar="HIDDEN_XYZ"
    )
    @click.option("--without-xyz", "enable_xyz", flag_value=False)
    @click.pass_context
    def cli(ctx, enable_xyz):
        src = ctx.get_parameter_source("enable_xyz")
        click.echo(f"value={enable_xyz!r} source={src.name}")

    result = CliRunner().invoke(cli, [], env={"HIDDEN_XYZ": "1"})
    assert result.exit_code == 0, result.output
    assert result.output.strip() == "value=True source=ENVIRONMENT"

    result = CliRunner().invoke(
        cli, [], env={"HIDDEN_XYZ": "1"}, default_map={"enable_xyz": False}
    )
    assert result.exit_code == 0, result.output
    assert result.output.strip() == "value=True source=ENVIRONMENT"

    result = CliRunner().invoke(cli, [], default_map={"enable_xyz": False})
    assert result.exit_code == 0, result.output
    assert result.output.strip() == "value=False source=DEFAULT_MAP"


def test_feature_switch_declaration_order_reversed():
    @click.command()
    @click.option("--with-xyz", "enable_xyz", flag_value=True, default=True)
    @click.option("--without-xyz", "enable_xyz", flag_value=False)
    @click.pass_context
    def cli(ctx, enable_xyz):
        src = ctx.get_parameter_source("enable_xyz")
        click.echo(f"value={enable_xyz!r} source={src.name}")

    result = CliRunner().invoke(cli, ["--without-xyz"], default_map={"enable_xyz": True})
    assert result.exit_code == 0, result.output
    assert result.output.strip() == "value=False source=COMMANDLINE"


def test_feature_switch_env_winner_other_declaration_order():
    @click.command()
    @click.option("--without-xyz", "enable_xyz", flag_value=False)
    @click.option(
        "--with-xyz", "enable_xyz", flag_value=True, default=True, envvar="HIDDEN_XYZ"
    )
    @click.pass_context
    def cli(ctx, enable_xyz):
        src = ctx.get_parameter_source("enable_xyz")
        click.echo(f"value={enable_xyz!r} source={src.name}")

    result = CliRunner().invoke(
        cli, [], env={"HIDDEN_XYZ": "1"}, default_map={"enable_xyz": False}
    )
    assert result.exit_code == 0, result.output
    assert result.output.strip() == "value=True source=ENVIRONMENT"
