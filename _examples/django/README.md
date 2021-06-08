# About django examples

The django examples are the result of running `xo` against all the supported
databases that django supports. It is helpful for comparing generated code
against the supported databases, as well because django has a non-trivial
schema with more realistic database relationships than the [simple booktest
examples](../booktest).

## Running

The [`gen.sh`](gen.sh) script will generate all the models for the supported
databases in each database's subfolder:

* [mysql](mysql/)
* [oracle](oracle/)
* [postgres](postgres/)
* [sqlite3](sqlite3/)
* [sqlserver](sqlserver/)

## Installing django

You can install/update to the latest version of Django on Debian/Ubuntu by
doing the following:

```sh
# install django globally
$ sudo pip3 install -U Django

# install support for sqlserver via odbc
$ sudo aptitude install unixodbc unixodbc-dev
$ sudo pip3 install -U django-pyodbc-azure
```
