from typing import Optional
from dataclasses import dataclass

@dataclass
class User:
    id: int
    name: str
    email: str

    def validate_email(self) -> bool:
        return "@" in self.email

    def to_dict(self) -> dict:
        return {
            "id": self.id,
            "name": self.name,
            "email": self.email
        }

class UserRepository:
    def __init__(self, db):
        self._db = db

    def find_by_id(self, user_id: int) -> Optional[User]:
        result = self._db.query(f"SELECT * FROM users WHERE id = {user_id}")
        if result:
            return User(**result)
        return None

    def save(self, user: User) -> bool:
        return self._db.execute("INSERT INTO users VALUES (?)", user.to_dict())
