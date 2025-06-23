<?php
// Test Pattern 5: Trait use Statements and Complex Patterns
// Tests trait imports, PHPDoc references, and global namespace fallback

namespace App\Models;

use App\Traits\SoftDeletes;
use App\Traits\Auditable;
use App\Traits\Searchable;
use App\Traits\Cacheable;
use App\Events\ModelCreated;
use App\Events\ModelUpdated;
use App\Events\ModelDeleted;
use App\Services\NotificationService;
use App\Services\CacheService;
use Illuminate\Database\Eloquent\Model;
use Carbon\Carbon;
use Ramsey\Uuid\Uuid;

// Not using: Auditable, Searchable, Cacheable, ModelUpdated, ModelDeleted, 
// CacheService, Model, Uuid

/**
 * Post model with soft deletes and notifications
 * 
 * @property int $id
 * @property string $title
 * @property string $content
 * @property Carbon $published_at
 * @property Carbon $created_at
 * @property Carbon $updated_at
 * 
 * @method static \Illuminate\Database\Eloquent\Builder query()
 * @see Model for base functionality
 */
class Post
{
    use SoftDeletes;
    // Could use other traits but choosing not to
    
    private NotificationService $notifier;
    private array $attributes = [];
    
    public function __construct(NotificationService $notifier)
    {
        $this->notifier = $notifier;
        $this->attributes['created_at'] = Carbon::now();
    }
    
    /**
     * Publish the post
     * 
     * @throws \Exception if already published
     * @see NotificationService::sendNotification()
     * @see ModelCreated for the event that's dispatched
     */
    public function publish(): void
    {
        if ($this->isPublished()) {
            // Using global namespace Exception (no import needed)
            throw new \Exception('Post is already published');
        }
        
        // Set published timestamp using Carbon
        $this->attributes['published_at'] = Carbon::now();
        
        // Use method from SoftDeletes trait
        if ($this->trashed()) {
            $this->restore();
        }
        
        // Send notification
        $this->notifier->sendNotification('post_published', [
            'id' => $this->attributes['id'],
            'title' => $this->attributes['title']
        ]);
        
        // Dispatch event
        event(new ModelCreated($this));
    }
    
    /**
     * Check if post is published
     * 
     * @return bool
     */
    public function isPublished(): bool
    {
        return isset($this->attributes['published_at']) 
            && $this->attributes['published_at'] instanceof Carbon;
    }
    
    /**
     * Get formatted dates
     * 
     * @param string $format
     * @return array
     * @see Carbon::format() for format options
     * @see Ramsey\Uuid\Uuid for ID generation (not used but referenced)
     */
    public function getFormattedDates(string $format = 'Y-m-d H:i:s'): array
    {
        $dates = [];
        
        // Format created_at if it exists and is Carbon instance
        if (isset($this->attributes['created_at']) 
            && $this->attributes['created_at'] instanceof Carbon) {
            $dates['created'] = $this->attributes['created_at']->format($format);
        }
        
        // Format published_at if it exists
        if ($this->isPublished()) {
            $dates['published'] = $this->attributes['published_at']->format($format);
        }
        
        return $dates;
    }
    
    /**
     * Magic method to access attributes
     * Uses global namespace without import
     */
    public function __get(string $name)
    {
        return $this->attributes[$name] ?? null;
    }
    
    public function __set(string $name, $value): void
    {
        $this->attributes[$name] = $value;
    }
}