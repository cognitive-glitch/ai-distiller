<?php
// Test Pattern 1: Basic Namespace use Statements
// Tests simple use statements for classes within namespaces

namespace App\Controllers;

use App\Models\User;
use App\Models\Product;
use App\Services\AuthService;
use App\Services\EmailService;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;
use Psr\Log\LoggerInterface;
use Doctrine\ORM\EntityManager;

// Not using: Product, EmailService, Request, LoggerInterface, EntityManager

class UserController
{
    private AuthService $authService;
    
    public function __construct(AuthService $authService)
    {
        $this->authService = $authService;
    }
    
    public function showUser(int $id): Response
    {
        // Using AuthService
        if (!$this->authService->isAuthenticated()) {
            return new Response('Unauthorized', 401);
        }
        
        // Using User model
        $user = User::find($id);
        if (!$user) {
            return new Response('User not found', 404);
        }
        
        // Using Response
        return new Response(json_encode([
            'id' => $user->getId(),
            'name' => $user->getName(),
            'email' => $user->getEmail()
        ]), 200, ['Content-Type' => 'application/json']);
    }
    
    public function listUsers(): Response
    {
        $users = User::all();
        return new Response(json_encode($users));
    }
}