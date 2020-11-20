#!/bin/bash

if [ "$1" = "register" ]; then
    sqlite ../backend/db.db "DELETE FROM users WHERE unique_id='88888888Y';"
fi