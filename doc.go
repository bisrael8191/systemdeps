/*
Utility library that manages systemd service dependencies using drop-in
configuration files instead of modifying the installed unit file.

Service dependencies can be stored in a JSON file so that it is similar
to defining services and depends_on links in docker compose.

For example, the base services can be installed using generic Ansible roles
then independently ordered per VM based on it's needs.

Optionally, can create a top-level application to manage all underlying processes
using standard systemd commands:
  systemctl [start | stop | restart | enable | disable] application.target

References

https://www.freedesktop.org/software/systemd/man/systemd.unit.html
https://www.freedesktop.org/software/systemd/man/systemd.target.html
*/
package systemdeps // import "github.com/bisrael8191/systemdeps"
