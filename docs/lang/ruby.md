# Ruby Language Support

AI Distiller provides comprehensive support for Ruby codebases using the [tree-sitter-ruby](https://github.com/tree-sitter/tree-sitter-ruby) parser, with full support for Ruby's dynamic nature, metaprogramming features, and modern syntax.

## Overview

Ruby support in AI Distiller captures the essential structure of Ruby code including classes, modules, methods, and constants. The distilled output preserves Ruby's object-oriented design and metaprogramming capabilities while optimizing for AI consumption.

## Recent Fixes (2025-06-15)

1. **Missing `def` keyword** (✅ Fixed)
   - **Issue**: Methods were displayed without the `def` keyword
   - **Fix**: Updated Ruby formatter to include proper method syntax
   - **Impact**: Correct Ruby method declarations in output

2. **Python-style colons** (✅ Fixed)
   - **Issue**: Classes and modules had Python-style colons after declarations
   - **Fix**: Removed colon from class/module formatting
   - **Impact**: Idiomatic Ruby syntax

## Supported Ruby Constructs

### Core Language Features

| Construct | Support Level | Notes |
|-----------|--------------|-------|
| **Classes** | ✅ Full | Including nested classes |
| **Modules** | ✅ Full | Including module nesting |
| **Methods** | ✅ Full | Instance, class methods (self.), module methods |
| **Attr Accessors** | ✅ Full | attr_reader, attr_writer, attr_accessor |
| **Constants** | ✅ Full | Class and module constants |
| **Instance Variables** | ✅ Full | @instance_var |
| **Class Variables** | ✅ Full | @@class_var |
| **Singleton Methods** | ✅ Full | def self.method |
| **Private/Protected** | ✅ Full | Visibility modifiers |
| **Include/Extend** | ✅ Full | Module mixins |
| **Alias** | ✅ Full | Method aliasing |
| **Inheritance** | ✅ Full | Class < SuperClass |
| **Blocks** | ⚠️ Partial | Block parameters in method signatures |
| **Metaprogramming** | ⚠️ Partial | define_method, method_missing |

### Visibility Rules

Ruby visibility in AI Distiller:
- **public**: Default visibility (not marked)
- **private**: Methods below `private` keyword
- **protected**: Methods below `protected` keyword
- **module_function**: Module methods that become instance methods

## Key Features

### 1. **Method Visibility**

AI Distiller correctly handles Ruby's visibility modifiers:

```ruby
// Input
class User
  def public_method
    "public"
  end
  
  private
  
  def private_method
    "private"
  end
  
  protected
  
  def protected_method
    "protected"
  end
end
```

```
// Output (default - public only)
class User
  def public_method
end
```

### 2. **Module Mixins**

Includes and extends are preserved:

```ruby
// Input
module Trackable
  def track(event)
    # tracking logic
  end
end

class Order
  include Trackable
  extend ActiveModel::Naming
  
  def process
    track(:order_processed)
  end
end
```

```
// Output
module Trackable
  def track(event)
end

class Order
  include Trackable
  extend ActiveModel::Naming
  
  def process
end
```

### 3. **Class Methods**

Different ways of defining class methods are supported:

```ruby
// Input
class Calculator
  def self.add(a, b)
    a + b
  end
  
  class << self
    def multiply(a, b)
      a * b
    end
  end
end
```

```
// Output
class Calculator
  def self.add(a, b)
  def self.multiply(a, b)
end
```

### 4. **Attr Accessors**

Ruby's attr_* methods are properly displayed:

```ruby
// Input
class Product
  attr_reader :id, :name
  attr_writer :price
  attr_accessor :quantity
  
  def initialize(id, name)
    @id = id
    @name = name
  end
end
```

```
// Output
class Product
  attr_reader :id, :name
  attr_writer :price
  attr_accessor :quantity
  
  def initialize(id, name)
end
```

## Output Format

### Text Format (Recommended for AI)

The text format preserves idiomatic Ruby syntax:

```ruby
// Input file
module Authentication
  extend ActiveSupport::Concern
  
  included do
    before_action :authenticate_user!
  end
  
  def authenticate_user!
    redirect_to login_path unless current_user
  end
  
  def current_user
    @current_user ||= User.find_by(id: session[:user_id])
  end
  
  private
  
  def set_current_user(user)
    @current_user = user
    session[:user_id] = user.id
  end
end

class ApplicationController < ActionController::Base
  include Authentication
  
  protect_from_forgery with: :exception
  
  rescue_from ActiveRecord::RecordNotFound do |e|
    render_404
  end
  
  private
  
  def render_404
    render file: 'public/404.html', status: :not_found
  end
end

// Output (default - public only, no implementation)
<file path="auth.rb">
module Authentication
  extend ActiveSupport::Concern
  
  def authenticate_user!
  def current_user
end

class ApplicationController < ActionController::Base
  include Authentication
end
</file>
```

## Known Limitations

1. **Block Syntax**: Complex block parameters may not be fully captured
2. **Metaprogramming**: Dynamic method definitions are not expanded
3. **String Interpolation**: Not parsed within method implementations
4. **Heredocs**: May not be properly formatted
5. **Lambda/Proc**: Type information limited

## Best Practices

### 1. **Use Explicit Visibility**

Group methods by visibility:

```ruby
class Service
  # Public interface
  def call(params)
    validate(params)
    process(params)
  end
  
  private
  
  # Implementation details
  def validate(params)
    # ...
  end
  
  def process(params)
    # ...
  end
end
```

### 2. **Module Organization**

Use modules for shared behavior:

```ruby
module Concerns
  module Searchable
    extend ActiveSupport::Concern
    
    included do
      scope :search, ->(query) { where("name LIKE ?", "%#{query}%") }
    end
    
    class_methods do
      def searchable_fields
        %w[name description]
      end
    end
  end
end
```

### 3. **Documentation**

Use YARD-style comments for better AI understanding:

```ruby
# Processes payment for an order
# @param order [Order] the order to process
# @param payment_method [String] the payment method to use
# @return [PaymentResult] the result of the payment
def process_payment(order, payment_method)
  # ...
end
```

## Integration Examples

### Direct CLI Usage

```bash
# Extract public API
aid app/ --format text --private=0 --protected=0 --internal=0 > api.txt

# Include all methods
aid app/ --format text --output full.txt

# Rails models only
aid app/models --format text --implementation=0 > models.txt
```

### Rails Projects

```bash
# Controllers API
aid app/controllers --private=0 --protected=0 --internal=0 --implementation=0 > controllers-api.txt

# Models with associations
aid app/models --format text > models.txt

# Service objects
aid app/services --format text --implementation=0 > services.txt
```

## Future Improvements

- Full block syntax support
- Metaprogramming expansion
- Better heredoc handling
- DSL recognition (RSpec, Rails, etc.)
- Refinements support

## Contributing

Ruby support is actively maintained. Key areas for contribution:
- Complex metaprogramming patterns
- DSL parsing (Rails, RSpec, etc.)
- Block and proc handling
- Performance optimizations

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for development setup.

---

<sub>Documentation generated for AI Distiller v0.2.0</sub>