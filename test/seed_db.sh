#!/bin/bash

if [ "$1" = "register" ]; then
    sqlite3 ../backend/db.db "DELETE FROM users WHERE unique_id='88888888Y';"
fi