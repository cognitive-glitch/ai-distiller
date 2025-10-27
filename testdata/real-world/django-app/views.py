"""
Django-style views with decorators and async support.
Tests: Decorators, async/await, error handling, HTTP patterns.
"""
from typing import Dict, Any, Optional
from functools import wraps


def login_required(func):
    """Decorator requiring user authentication."""
    @wraps(func)
    def wrapper(request, *args, **kwargs):
        if not request.user.is_authenticated:
            return {"error": "Authentication required"}, 401
        return func(request, *args, **kwargs)
    return wrapper


def api_view(methods: list):
    """Decorator specifying allowed HTTP methods."""
    def decorator(func):
        @wraps(func)
        def wrapper(request, *args, **kwargs):
            if request.method not in methods:
                return {"error": "Method not allowed"}, 405
            return func(request, *args, **kwargs)
        wrapper.allowed_methods = methods
        return wrapper
    return decorator


class Request:
    """Mock request object."""
    def __init__(self, method: str, user: Any):
        self.method = method
        self.user = user


@api_view(['GET', 'POST'])
@login_required
def user_list_view(request: Request) -> tuple[Dict[str, Any], int]:
    """Handle user list requests."""
    if request.method == 'GET':
        return {"users": []}, 200
    elif request.method == 'POST':
        return {"created": True}, 201
    return {"error": "Invalid method"}, 400


@api_view(['GET'])
def user_detail_view(request: Request, user_id: int) -> tuple[Dict[str, Any], int]:
    """Get user details."""
    if user_id <= 0:
        return {"error": "Invalid user ID"}, 400
    return {"user": {"id": user_id}}, 200


async def async_user_view(request: Request) -> Dict[str, Any]:
    """Async view for user data."""
    # Simulate async database call
    return {"users": []}


class UserViewSet:
    """Class-based view for user operations."""
    
    def list(self, request: Request) -> Dict[str, Any]:
        """List all users."""
        return {"users": []}
    
    def retrieve(self, request: Request, pk: int) -> Dict[str, Any]:
        """Get single user."""
        return {"user": {"id": pk}}
    
    def create(self, request: Request, data: Dict[str, Any]) -> Dict[str, Any]:
        """Create new user."""
        return {"created": True, "data": data}
    
    def update(self, request: Request, pk: int, data: Dict[str, Any]) -> Dict[str, Any]:
        """Update user."""
        return {"updated": True, "id": pk}
    
    def destroy(self, request: Request, pk: int) -> Dict[str, Any]:
        """Delete user."""
        return {"deleted": True, "id": pk}
