package ysql

// mimic how pgx does QueryRow except we're not bothering with an interface
type Row Rows

func (r *Row) Scan(dest ...interface{}) error {
	rows := (*Rows)(r)

	if rows.Err() != nil {
		return rows.Err()
	}

	if !rows.Next() {
		if rows.Err() == nil {
			return ErrNoRows
		}
		return rows.Err()
	}

	err := rows.Scan(dest...)
	rows.Close()
	if err != nil {
		return err
	}
	return rows.Err()
}
