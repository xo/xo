# the following needs to be copied/pasted into the settings.py or __init__.py
# file for django to correctly generate its sqlite3 database (enabling foreign
# keys).
#
# see: http://stackoverflow.com/questions/6745763/enable-integrity-checking-with-sqlite-in-django/6835016

from django.db.backends.signals import connection_created

def activate_foreign_keys(sender, connection, **kwargs):
    """Enable integrity constraint with sqlite."""
    if connection.vendor == 'sqlite':
        cursor = connection.cursor()
        cursor.execute('PRAGMA foreign_keys = ON;')

connection_created.connect(activate_foreign_keys)
