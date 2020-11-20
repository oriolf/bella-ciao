#!/bin/bash

db="../backend/db.db"
if [ "$1" = "register" ]; then
    sqlite3 $db "DELETE FROM users WHERE unique_id='88888888Y';"
elif [ "$1" = "validate" ]; then
    sqlite3 $db "UPDATE users SET role='none' WHERE unique_id='88888888Y';"
elif [ "$1" = "vote" ]; then
    datetime=$(date -d '1 hour ago' --rfc-3339='ns')
    sqlite3 $db "UPDATE users SET has_voted=0 WHERE unique_id='88888888Y';"
    sqlite3 $db "DELETE FROM votes;"
    sqlite3 $db "UPDATE elections SET date_start='$datetime', counted=0;"
fi
