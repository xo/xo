# Templates

Templates are separated into **Languages**, **Drivers** and **Protocols**.

```
Languages
-- Drivers
-- Protocols
```

## Language

The repository provides template generation for all languages: A directory named after a programming language (i.e go) will contain templated driver code.

### Drivers

A driver represents a database technology or type that a template provides code for. In the context of the library, protocols involve generating database code from the database schema. For example, SQL, PSQL, and Oracle are all database technologies which use separate drivers.

### Protocol

A protocol represents a form of data acess that uses a schema. For example, SQL, REST, gRPC, etc. In the context of the library, protocols involve generating database code from the protocol schema and/or connecting them to drivers.

