"""Create or update a user in the SQLite user store.

Usage:
    python seed_user.py --username admin
(You'll be prompted for a password; it's never echoed or logged.)
"""
import argparse
import getpass

import auth_db


def main():
    parser = argparse.ArgumentParser(description="Create or update a login user")
    parser.add_argument("--username", required=True)
    args = parser.parse_args()

    password = getpass.getpass("Password: ")
    confirm = getpass.getpass("Confirm password: ")
    if password != confirm:
        print("Passwords do not match.")
        raise SystemExit(1)
    if len(password) < 8:
        print("Password must be at least 8 characters.")
        raise SystemExit(1)

    auth_db.upsert_user(args.username, password)
    print(f"User '{args.username}' created/updated.")


if __name__ == "__main__":
    main()
