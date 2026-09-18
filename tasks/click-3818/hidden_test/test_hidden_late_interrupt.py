"""Hidden test: a KeyboardInterrupt arriving while Command.main() reports the
outcome must not escape and must not change the exit code."""

import sys

import pytest

import click


class InterruptingStderr:
    """stderr stand-in; raises KeyboardInterrupt on the N-th isatty() or
    write() call (counted together), if ``trigger`` is set."""

    def __init__(self, trigger=None):
        self.trigger = trigger
        self.calls = 0
        self.text = ""

    def _tick(self):
        self.calls += 1
        if self.calls == self.trigger:
            raise KeyboardInterrupt

    def isatty(self):
        self._tick()
        return False

    def write(self, s):
        self._tick()
        self.text += s
        return len(s)

    def flush(self):
        pass


def _run_standalone(cli, monkeypatch, stderr):
    monkeypatch.setattr(sys, "stderr", stderr)
    try:
        cli.main([], "prog")
    except SystemExit as e:
        return e.code
    except KeyboardInterrupt:
        pytest.fail("KeyboardInterrupt escaped Command.main()")
    pytest.fail("Command.main() returned instead of exiting in standalone mode")


def _prompt_cli():
    @click.command()
    @click.option("--name", prompt="Name")
    def cli(name):
        click.echo(name)

    return cli


def _body_interrupt_cli(exc=KeyboardInterrupt):
    @click.command()
    def cli():
        raise exc

    return cli


def test_single_interrupt_still_prints_aborted(monkeypatch):
    monkeypatch.setattr("click.termui.visible_prompt_func", _raise_interrupt)
    stderr = InterruptingStderr()
    assert _run_standalone(_prompt_cli(), monkeypatch, stderr) == 1
    assert "Aborted!" in stderr.text


def _raise_interrupt(prompt):
    raise KeyboardInterrupt


@pytest.mark.parametrize("trigger", [1, 2, 3, 4])
def test_late_interrupt_during_abort_from_prompt(monkeypatch, trigger):
    monkeypatch.setattr("click.termui.visible_prompt_func", _raise_interrupt)
    stderr = InterruptingStderr(trigger=trigger)
    assert _run_standalone(_prompt_cli(), monkeypatch, stderr) == 1


@pytest.mark.parametrize("exc", [KeyboardInterrupt, EOFError])
@pytest.mark.parametrize("trigger", [1, 2, 3, 4, 5])
def test_late_interrupt_during_abort_from_body(monkeypatch, exc, trigger):
    stderr = InterruptingStderr(trigger=trigger)
    assert _run_standalone(_body_interrupt_cli(exc), monkeypatch, stderr) == 1


def test_late_interrupt_while_showing_click_exception(monkeypatch):
    class Boom(click.ClickException):
        exit_code = 7

        def show(self, file=None):
            raise KeyboardInterrupt

    @click.command()
    def cli():
        raise Boom("boom")

    assert _run_standalone(cli, monkeypatch, InterruptingStderr()) == 7


def test_late_interrupt_while_writing_usage_error(monkeypatch):
    @click.command()
    def cli():
        raise click.UsageError("bad usage")

    # UsageError.show writes several chunks to stderr; interrupt mid-way.
    for trigger in (1, 2, 3):
        stderr = InterruptingStderr(trigger=trigger)
        assert _run_standalone(cli, monkeypatch, stderr) == 2


def test_late_interrupt_while_exiting_after_success(monkeypatch):
    @click.command()
    def cli():
        click.echo("done")

    real_exit = sys.exit
    seen = []

    def exit_then_interrupt(code=None):
        seen.append(code)
        if len(seen) == 1:
            raise KeyboardInterrupt
        real_exit(code)

    monkeypatch.setattr(sys, "exit", exit_then_interrupt)
    assert _run_standalone(cli, monkeypatch, InterruptingStderr()) == 0
    assert all(code in (0, None) for code in seen)


def test_non_standalone_first_interrupt_is_abort(monkeypatch):
    monkeypatch.setattr(sys, "stderr", InterruptingStderr())
    with pytest.raises(click.Abort) as info:
        _body_interrupt_cli().main([], "prog", standalone_mode=False)
    assert isinstance(info.value.__cause__, KeyboardInterrupt)
