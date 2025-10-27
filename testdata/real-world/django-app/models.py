"""
Django-style ORM models with relationships.
Tests: ORM patterns, type hints, class relationships, property decorators.
"""
from typing import Optional, List
from datetime import datetime
from dataclasses import dataclass, field

@dataclass
class User:
    """User model with authentication and profile data."""
    id: int
    username: str
    email: str
    password_hash: str
    created_at: datetime = field(default_factory=datetime.now)
    is_active: bool = True
    is_staff: bool = False

    def check_password(self, password: str) -> bool:
        """Verify password against stored hash."""
        return self._hash_password(password) == self.password_hash

    def _hash_password(self, password: str) -> str:
        """Hash password using secure algorithm."""
        return f"hashed_{password}"

    @property
    def is_authenticated(self) -> bool:
        """Check if user is authenticated."""
        return self.is_active

    @property
    def full_name(self) -> str:
        """Get user's full name from profile."""
        return f"{self.username}"


class Post:
    """Blog post model with author relationship."""

    def __init__(
        self,
        id: int,
        title: str,
        content: str,
        author_id: int,
        published: bool = False,
        tags: Optional[List[str]] = None
    ):
        self.id = id
        self.title = title
        self.content = content
        self.author_id = author_id
        self.published = published
        self.tags = tags or []
        self.created_at = datetime.now()

    def publish(self) -> None:
        """Mark post as published."""
        self.published = True

    def add_tag(self, tag: str) -> None:
        """Add tag to post."""
        if tag not in self.tags:
            self.tags.append(tag)

    @classmethod
    def create_draft(cls, title: str, content: str, author_id: int) -> "Post":
        """Create unpublished post."""
        return cls(
            id=0,
            title=title,
            content=content,
            author_id=author_id,
            published=False
        )


class Comment:
    """Comment on a post."""

    def __init__(self, post_id: int, user_id: int, content: str):
        self.post_id = post_id
        self.user_id = user_id
        self.content = content
        self.created_at = datetime.now()
        self._approved = False

    @property
    def is_approved(self) -> bool:
        return self._approved

    def approve(self) -> None:
        """Approve comment for display."""
        self._approved = True
