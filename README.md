# ysql
Make composing string SQL for github.com/jackc/pgx/v4 a little easier

This introduces a couple of new features to the fantastic github.com/jackc/pgx/v4.

1) You can Scan() into a struct and it will match by field name (or struct tag `ysql:"field_name"`).
  There are some other cool packages that do this already (https://github.com/jmoiron/sqlx or https://github.com/georgysavva/scany).
  so this isn't really anything special and I'm probably not as optimized as those other packages are.

2) Some magical placeholder stuff!

Instead of just $1, $2, etc, you can use a field name:

    select first_name from users where id = $id;

This will expand to the value in the struct for the field that matches that name.
Another shortcut is for when you need field = $field, which is very common in selects and updates:

    select first_name from users where $=id;

And my favorite shortcut is for inserts, which is a sort of meta-splat operator. $$field will be replaced by the name of the field, 
and $... will be replaced by a comma separated list of the field values.

    insert into users ($$first_name, $$last_name) values ($...);

Example:

    type User struct {
        ID int
        FirstName string `ysql:"first_name"`
        LastName string `ysql:"last_name"`
    }
    
    var user User
    _, err := ysql.Exec(conn, ctx, `update users set first_name = $first_name, last_name = $last_name where id = $id;`, user);
    // the exact same as:
    _, err := ysql.Exec(conn, ctx, `update users set $=first_name, $=last_name where $=id;`, user);
    
    // splat syntax:
    _, err := ysql.Exec(conn, ctx, `insert into users ($$first_name, $$last_name) values ($...);`, user);
    
    // reading a row: also traditional numbered placeholders still work as you'd expect
    err := ysql.QueryRow(conn, ctx, `select first_name, last_name from users where id = $1;`, id).Scan(&user)
    
    // reading many rows: 
    rows, err := ysql.QueryRow(conn, ctx, `select first_name from users where $=first_name;`, user)
    for rows.Next() {
        var u User
        err := rows.Scan(&u)
        // etc
    }
