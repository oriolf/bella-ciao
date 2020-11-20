#!/bin/bash

if [ "$1" = "register" ]; then
    sqlite3 ../backend/db.db "DELETE FROM users WHERE unique_id='88888888Y';"
elif [ "$1" = "validate" ]; then
    sqlite3 ../backend/db.db "UPDATE users SET role='none' WHERE unique_id='88888888Y';"
fi
