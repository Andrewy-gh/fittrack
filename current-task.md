I want to remove all `SELECT *` queries from the database as it is not secure and can be a performance issue.

Can we check all frontend components that are using queries with `SELECT *`?

For each component that is using a `SELECT *` query, check exactly with columns are needed and replace the query with the appropriate query.

Once we have a good idea of the queries that need to be changed, we can start working on replacing the return values of the queries.