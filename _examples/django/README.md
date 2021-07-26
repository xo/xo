# About django examples

The `django` example is the result of running `xo` against all the supported
databases that Django and `xo` supports, with Django models similar to the `xo`
booktest schema.

## Setup

Install packages:

```sh
# install mysql, postgres, sqlite3 dependencies
$ sudo aptitude install libpq-dev libmysqlclient-dev libsqlite3-dev

# install sqlserver dependenices
# manually add the microsoft-prod ppa
$ sudo aptitude install unixodbc-dev msodbcsql17

# ensure odbcinst.ini has the relevant sqlserver entry
$ cat /etc/odbcinst.ini
[ODBC Driver 17 for SQL Server]
Description=Microsoft ODBC Driver 17 for SQL Server
Driver=/opt/microsoft/msodbcsql17/lib64/libmsodbcsql-17.7.so.2.1
UsageCount=1

# install oracle dependencies
$ cd /path/to/usql/contrib/godror
$ sudo ./grab-instantclient.sh

# install pipenv
$ pip install --user pipenv

# install packages
$ pipenv install

# update packages
$ pipenv update
```
