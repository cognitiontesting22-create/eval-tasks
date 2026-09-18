"""Hidden test: a user parameter named ``help`` must coexist with the automatic
help option."""

import click
from click.testing import CliRunner


def _runner():
    return CliRunner()


def test_argument_named_help_receives_value():
    @click.command()
    @click.argument("help")
    def cli(help):
        click.echo(f"got={help}")

    result = _runner().invoke(cli, ["this"])
    assert result.exit_code == 0, result.output
    assert result.output == "got=this\n"


def test_argument_named_help_keeps_help_flag():
    @click.command()
    @click.argument("help")
    def cli(help):
        click.echo(f"got={help}")

    result = _runner().invoke(cli, ["--help"])
    assert result.exit_code == 0, result.output
    assert "Show this message and exit." in result.output
    assert "got=" not in result.output
    assert "Missing argument" not in result.output


def test_option_with_destination_named_help():
    @click.command()
    @click.option("--assist", "help")
    def cli(help):
        click.echo(f"got={help}")

    result = _runner().invoke(cli, ["--assist", "value"])
    assert result.exit_code == 0, result.output
    assert result.output == "got=value\n"

    result = _runner().invoke(cli, ["--help"])
    assert result.exit_code == 0, result.output
    assert "Show this message and exit." in result.output
    assert "--assist" in result.output


def test_option_named_help_with_default():
    @click.command()
    @click.option("--assist", "help", default="fallback")
    def cli(help):
        click.echo(f"got={help}")

    result = _runner().invoke(cli, [])
    assert result.exit_code == 0, result.output
    assert result.output == "got=fallback\n"


def test_custom_help_option_name_conflict():
    @click.command(context_settings={"help_option_names": ["--man"]})
    @click.option("--foo", "man")
    def cli(man):
        click.echo(f"got={man}")

    result = _runner().invoke(cli, ["--foo", "value"])
    assert result.exit_code == 0, result.output
    assert result.output == "got=value\n"

    result = _runner().invoke(cli, ["--man"])
    assert result.exit_code == 0, result.output
    assert "Show this message and exit." in result.output


def test_argument_named_help_in_group_subcommand():
    @click.group()
    def cli():
        pass

    @cli.command()
    @click.argument("help")
    def show(help):
        click.echo(f"got={help}")

    result = _runner().invoke(cli, ["show", "thing"])
    assert result.exit_code == 0, result.output
    assert result.output == "got=thing\n"

    result = _runner().invoke(cli, ["show", "--help"])
    assert result.exit_code == 0, result.output
    assert "Show this message and exit." in result.output


def test_option_reusing_help_flag_still_takes_it_over():
    @click.command()
    @click.option("--help", default="this_2")
    def cli(help):
        click.echo(f"got={help}")

    result = _runner().invoke(cli, [])
    assert result.exit_code == 0, result.output
    assert result.output == "got=this_2\n"

    result = _runner().invoke(cli, ["--help", "custom"])
    assert result.exit_code == 0, result.output
    assert result.output == "got=custom\n"


def test_plain_command_help_unchanged():
    @click.command()
    @click.argument("name")
    def cli(name):
        click.echo(name)

    result = _runner().invoke(cli, ["--help"])
    assert result.exit_code == 0
    assert "--help" in result.output
    assert "Show this message and exit." in result.output
