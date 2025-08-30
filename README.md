# homelab

# Remote Development Setup

This project is configured for remote development using VS Code and the Remote - SSH extension. Code is edited locally but lives and runs on your Linux machine.

## Getting Started

1. Install the [Remote - SSH extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-ssh) in VS Code.
2. Ensure you have SSH access to your Linux machine.
3. Update `.vscode/settings.json` if your username is not `jweber`:
	- Change `/home/jweber/projects/homelab` to your actual home directory path.
4. Open the folder in VS Code using "Remote-SSH: Connect to Host...".

## Recommended Extensions

Extensions are listed in `.vscode/extensions.json`. VS Code will prompt you to install them when you open the project.

## Notes

- Each user should update the workspace path in `.vscode/settings.json` to match their Linux username if different.

## PreRequisites

```bash
$ sudo dnf install ansible-core ansible-collection-ansible-posix
```

## Secrets and Vault Usage

You must create an Ansible vault file at `~/.homelab.vault` (outside this repository).

This vault file should contain at least the following variables:

- `ansible_user`: The username to SSH as for remote connections.
- `ansible_become_password`: The sudo password for privilege escalation.

Example vault file contents:

```yaml
ansible_user: your_ssh_username
ansible_become_password: your_sudo_password
```

This file is referenced in all playbooks via `vars_files: [ ~/.homelab.vault ]` and must exist for automation to work.
