package php

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/janreges/ai-distiller/internal/ir"
)

func TestProcessorBasicClass(t *testing.T) {
	source := `<?php
namespace App\Models;

use App\Core\BaseModel;

class User extends BaseModel
{
    public const TYPE_USER = 'user';
    
    public string $name;
    protected string $email;
    private string $password;
    
    public function __construct(string $name, string $email)
    {
        $this->name = $name;
        $this->email = $email;
    }
    
    public function getName(): string
    {
        return $this->name;
    }
    
    protected function getEmail(): string
    {
        return $this->email;
    }
    
    private function validatePassword(string $password): bool
    {
        return strlen($password) >= 8;
    }
}`

	proc := NewProcessor()
	proc.EnableTreeSitter()
	
	ctx := context.Background()
	file, err := proc.Process(ctx, strings.NewReader(source), "test.php")
	
	require.NoError(t, err)
	require.NotNil(t, file)
	
	// Check file metadata
	assert.Equal(t, "test.php", file.Path)
	assert.Equal(t, "php", file.Language)
	
	// Find the User class
	var userClass *ir.DistilledClass
	for _, child := range file.Children {
		if class, ok := child.(*ir.DistilledClass); ok {
			if class.Name == "User" {
				userClass = class
				break
			}
		}
	}
	
	require.NotNil(t, userClass, "User class not found")
}

func TestProcessorPHP8Features(t *testing.T) {
	source := `<?php
namespace App\Modern;

enum Status: string
{
    case PENDING = 'pending';
    case APPROVED = 'approved';
}

#[Route('/api/users')]
class UserController
{
    public function __construct(
        private readonly string $apiKey,
        public string|int $version = 1
    ) {}
    
    #[Route('/api/users/{id}', methods: ['GET'])]
    public function show(string|int $id): array
    {
        return ['id' => $id];
    }
}`

	proc := NewProcessor()
	proc.EnableTreeSitter()
	
	ctx := context.Background()
	file, err := proc.Process(ctx, strings.NewReader(source), "modern.php")
	
	require.NoError(t, err)
	require.NotNil(t, file)
	
	// Should parse without errors
	assert.Empty(t, file.Errors)
}

func TestProcessorNamespaceAndUse(t *testing.T) {
	source := `<?php
namespace App\Core\Services;

use App\Contracts\ServiceInterface;
use App\Traits\{Loggable, Cacheable};
use App\Events\UserCreated as UserCreatedEvent;
use function App\Helpers\sanitize_input;
use const App\Config\DEFAULT_TIMEOUT;

class UserService implements ServiceInterface
{
    use Loggable, Cacheable;
    
    public function createUser(): void
    {
        // Implementation
    }
}`

	proc := NewProcessor()
	proc.EnableTreeSitter()
	
	ctx := context.Background()
	file, err := proc.Process(ctx, strings.NewReader(source), "namespace.php")
	
	require.NoError(t, err)
	require.NotNil(t, file)
	
	// Check that imports are captured
	importCount := 0
	for _, child := range file.Children {
		if _, ok := child.(*ir.DistilledImport); ok {
			importCount++
		}
	}
	
	// Should have at least some imports
	assert.Greater(t, importCount, 0, "No imports found")
}

func TestProcessorTraits(t *testing.T) {
	source := `<?php
trait Loggable
{
    public function log(string $message): void
    {
        error_log($message);
    }
}

trait Timestampable
{
    public DateTime $createdAt;
    public DateTime $updatedAt;
}

class User
{
    use Loggable, Timestampable;
    
    public string $name;
}`

	proc := NewProcessor()
	proc.EnableTreeSitter()
	
	ctx := context.Background()
	file, err := proc.Process(ctx, strings.NewReader(source), "traits.php")
	
	require.NoError(t, err)
	require.NotNil(t, file)
	
	// Should have parsed traits and class
	assert.GreaterOrEqual(t, len(file.Children), 3, "Expected at least 3 top-level nodes")
}

func TestProcessorInterfaces(t *testing.T) {
	source := `<?php
interface Jsonable
{
    public function toJson(): string;
}

interface Arrayable
{
    public function toArray(): array;
}

interface Serializable extends Jsonable, Arrayable
{
    public function serialize(): string;
}

class Model implements Serializable
{
    public function toJson(): string
    {
        return '{}';
    }
    
    public function toArray(): array
    {
        return [];
    }
    
    public function serialize(): string
    {
        return '';
    }
}`

	proc := NewProcessor()
	proc.EnableTreeSitter()
	
	ctx := context.Background()
	file, err := proc.Process(ctx, strings.NewReader(source), "interfaces.php")
	
	require.NoError(t, err)
	require.NotNil(t, file)
	
	// Should have parsed interfaces and class
	assert.GreaterOrEqual(t, len(file.Children), 4, "Expected at least 4 top-level nodes")
}

func TestProcessorFallbackToLineBased(t *testing.T) {
	// Test that we can fall back gracefully
	proc := NewProcessor()
	proc.useTreeSitter = false // Force line-based parser
	
	source := `<?php
class User {
    public $name;
}`
	
	ctx := context.Background()
	file, err := proc.Process(ctx, strings.NewReader(source), "fallback.php")
	
	require.NoError(t, err)
	require.NotNil(t, file)
	
	// Should at least return a valid file structure
	assert.Equal(t, "php", file.Language)
	assert.Equal(t, "fallback.php", file.Path)
}