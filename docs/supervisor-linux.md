# Try the decision service on a Linux host

This guide installs the [decision service](./supervisor.md) on one Linux
host and walks one approval from the agent's request to the password of
the approver. It uses two accounts: your own account, which runs the
agent, and a second account, which approves. Use a test machine or a
virtual machine. The install changes system files for every user of the
host.

For a fleet, the [managed install guide](./mdm.md) has the contract and
the MDM recipes.

## What you need

- A Linux host with systemd and a system bus.
- polkit, with `pkttyagent`.
- Go and `make`, to build Gryph from source.
- A way to log in as a second account: a second virtual terminal, or SSH.

Install the packages:

| Distribution | Command |
|---|---|
| Arch Linux | `sudo pacman -S --needed polkit dbus go make` |
| Debian, Ubuntu | `sudo apt install dbus polkitd pkexec golang make` |
| Fedora | `sudo dnf install polkit dbus golang make` |

polkit starts on demand. `systemctl status polkit` can show `inactive`
before the first request.

## 0. Check that the test names are free

The guide makes an account, a group and files with fixed names, and the
teardown removes them. Check that none of them exists before you start:

```sh
id approver
getent group gryph-admins
ls -d /etc/safedep/gryph /opt/safedep/gryph /usr/bin/gryph /root/gryph-test
```

Each command must report that the name does not exist. If an account or a
group exists, use another name in every step. If `/root/gryph-test`
exists, use another directory for the input files in steps 3, 4 and 10. If Gryph is already
installed from a package, skip step 1 and keep the package files in the
teardown.

## 1. Build and place the binary

The managed install accepts only a binary on a path that root owns from
the file up to `/`. A binary in your home fails with exit 3.

```sh
git clone https://github.com/safedep/gryph && cd gryph
make gryph
sudo install -D -m 755 bin/gryph /opt/safedep/gryph/bin/gryph
sudo ln -sf /opt/safedep/gryph/bin/gryph /usr/bin/gryph
```

## 2. Make the approver account

A member of the admin group answers requests. Make the group and a second
account in it, with a password.

```sh
sudo groupadd gryph-admins
sudo useradd -m -s /bin/bash -G gryph-admins approver
sudo passwd approver
```

Your own account must not be in the group. The account that asks can
never answer its own request.

## 3. Write the managed configuration and the policy

The input files must pass the same root-owned path check as the binary,
so keep them under a directory that root owns.

```sh
sudo mkdir -p /root/gryph-test
sudo tee /root/gryph-test/managed.yml >/dev/null <<'YAML'
policy:
  enabled: true
  log_all_evaluations: true
  approval:
    channels: [same-user-tty, local-admin]
    min_assurance: local-admin
    inline_wait: 15s
    request_ttl: 30m
    max_grant_scope: session
    local_admin:
      group: gryph-admins
      allow_self_elevated: false
      allow_without_auth: false
managed:
  agents: [claude-code]
supervisor:
  enabled: true
YAML
sudo tee /root/gryph-test/policy.yaml >/dev/null <<'YAML'
version: "1"
rules:
  - id: deploy-needs-admin
    action: escalate
    severity: high
    match:
      action_types: [command_exec]
      command_patterns:
        - "make deploy"
    message: "a deploy needs an admin"
YAML
```

What these keys do in this test:

- `min_assurance: local-admin` stops you from approving your own request
  on your terminal.
- `allow_without_auth: false` makes the service refuse an answer without
  the password of the approver.
- `allow_self_elevated: false` makes the service refuse an answer from
  your own login session through `su` or `sudo`.

The [approval workflow](./security-policy.md#a-managed-host-with-the-decision-service)
has every key.

## 4. Install

```sh
sudo gryph install --managed \
  --config /root/gryph-test/managed.yml \
  --policy /root/gryph-test/policy.yaml --json
```

The install does these steps in order:

1. It writes the managed files to `/etc/safedep/gryph/`.
2. It makes the `_gryph` service account.
3. It writes the systemd units and the polkit action file
   `/usr/share/polkit-1/actions/io.safedep.gryph.policy`.
4. It starts the socket and waits for the service to answer.
5. It turns the switch on.

The report ends with `"status": "ok"`.

## 5. Check the install

```sh
sudo gryph doctor --managed
systemctl status gryph-supervisor.socket gryph-supervisor.service
pkaction --action-id io.safedep.gryph.approve --verbose
gryph doctor
```

Expected:

- `gryph doctor --managed` reports the `locked` profile and
  `approvers same-user-tty, local-admin (group gryph-admins)`.
- The socket is `active (listening)` and the service is `active (running)`.
- `pkaction` shows `auth_self` for `implicit any`, `implicit inactive` and
  `implicit active`.
- `gryph doctor` as your own account shows the row `Decision service`
  with `/run/safedep/gryph/hook.sock, profile enforce`.

## 6. Make a request as the developer

As your own account, start Claude Code in a project and ask it to run
`make deploy`. The hook blocks the command, and the agent reads:

```
This action needs approval. Request 7f3c2a1b is pending. Tell the user. ...
```

Without Claude Code, send the same hook event by hand. This is the
command the managed hook entry runs:

```sh
cat > /tmp/deploy.json <<'JSON'
{"session_id": "test-1", "cwd": "/tmp", "hook_event_name": "PreToolUse",
 "tool_name": "Bash", "tool_input": {"command": "make deploy"},
 "tool_use_id": "t1"}
JSON
gryph _hook claude-code PreToolUse < /tmp/deploy.json
```

The command exits non-zero and prints the same message on stderr. Note the
request ID.

## 7. Answer as the approver

Log in as `approver` from a new login: a second virtual terminal
(Ctrl+Alt+F3) or `ssh approver@localhost`. Do not use `su` from your own
terminal. That keeps your login session, and the service treats the
answer as your own (step 8).

```sh
gryph policy approve list
gryph policy approve resolve --id 7f3c2a1b --decision allow --scope session --note "test"
```

`resolve` asks for a confirmation on the terminal, then polkit asks for
the password of `approver`:

```
Answer request 7f3c2a1b with ALLOW? [y/N]: y
==== AUTHENTICATING FOR io.safedep.gryph.approve ====
Authenticate to answer the approval request
Authenticating as: approver
Password:
==== AUTHENTICATION COMPLETE ====
Request 7f3c2a1b is approved (local-auth), grant 4c1d9e20
```

`local-auth` is the assurance: the password proves a person at the
keyboard. `gryph policy approve show 7f3c2a1b` shows the state, the
assurance, the approver and the scope. In a desktop session, the polkit
agent of the desktop can show a dialog instead of the terminal prompt.

## 8. Check what the service refuses

Make a new request (step 6), then try each answer below. Each one leaves
the request pending.

| Try | Result |
|---|---|
| A wrong password, or Ctrl+C at the password prompt | `the answer needs the password of the approver: ...` |
| `resolve` as your own account | `the account that asked cannot answer its own request` |
| `su - approver` from your own terminal, then `resolve` | `the answer comes from the login session that asked (self-elevated), ...` |
| `resolve --yes` | refused on a managed host |

## 9. Retry as the developer

Ask the agent to run `make deploy` again, or send the hook event again:

```sh
gryph _hook claude-code PreToolUse < /tmp/deploy.json
```

The command passes on the session grant. The first hook after the answer
also prints `Request 7f3c2a1b was approved. You can retry.` Read the
record:

```sh
gryph policy approve history --format json
gryph policy receipts
```

Each row names the channel, the assurance, the approver and the grant.
The receipts carry the signature of the machine key.

## 10. Change the policy

Edit `/root/gryph-test/policy.yaml` and run the install command of step 4
again. The install writes the difference and reloads the service. To
reload by hand after an edit of `/etc/safedep/gryph/policy.yaml`:

```sh
sudo systemctl reload gryph-supervisor
```

The service keeps the group check of an account for 30 seconds. If an
account tried to answer before you added it to the group, it can answer
again after that time.

## When something does not work

```sh
sudo journalctl -u gryph-supervisor -f
sudo journalctl -u polkit -f
sudo gryph doctor --managed --json
```

| Symptom | Cause |
|---|---|
| The install exits 3 with a chain error | A file or a directory above the binary or the input files is not owned by root, or is writable by group or other. |
| `no approver channel` in `gryph doctor --managed` | The configuration names no `policy.approval.local_admin.group` for a `local-admin` or `self-elevated` channel, and turns on neither `local-auth` nor `out-of-band`. |
| `... no password was given: ...` | `resolve` ran without a terminal and without a polkit agent. Run it on a terminal. |
| `... the authority of the host stopped answering: ...` | polkit or the system bus stopped during the answer. |
| `resolve` reports `(local-admin)`, not `(local-auth)` | The service found no polkit on the system bus and took the answer without a password. Start polkit and check the action file (step 5). |
| The hook blocks every action with the service down | This is the `enforce` profile. See [fail modes](./supervisor.md#fail-modes). |

## Remove the test

These commands remove only what this guide made, if step 0 found none of
the names.

```sh
sudo gryph uninstall --managed --purge --json
sudo userdel -r approver
sudo groupdel gryph-admins
sudo rm -rf /root/gryph-test /opt/safedep/gryph /usr/bin/gryph
```

`uninstall --managed` turns the switch off first, then removes the units,
the polkit action file, the hook entries and the managed files.
