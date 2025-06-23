# Test Pattern 5: Extend, prepend, and metaprogramming
# Tests extend/prepend module usage and dynamic requires

require 'benchmark'
require 'delegate'
require 'observer'
require 'set'

# Autoload for lazy loading
autoload :ExpensiveModule, 'expensive_module'
autoload :RarelyUsedClass, 'rarely_used_class'

# Modules for extension and prepending
module ClassMethods
  def class_attribute(name)
    define_method(name) do
      instance_variable_get("@#{name}")
    end
    
    define_method("#{name}=") do |value|
      instance_variable_set("@#{name}", value)
    end
  end
end

module InstanceMethods
  def benchmark_method(method_name)
    # Using Benchmark
    result = nil
    time = Benchmark.measure do
      result = send(method_name)
    end
    puts "#{method_name} took #{time.real} seconds"
    result
  end
end

module Overrides
  def save
    puts "Before save hook"
    super
    puts "After save hook"
  end
end

module Observable
  # Including the standard Observable module
  include ::Observable
  
  def notify_change(attribute, value)
    changed
    notify_observers(attribute, value)
  end
end

# Not using: delegate, set, ExpensiveModule (autoloaded), RarelyUsedClass

class BaseModel
  extend ClassMethods
  include InstanceMethods
  
  class_attribute :table_name
  class_attribute :primary_key
end

class User < BaseModel
  prepend Overrides
  include Observable
  
  self.table_name = 'users'
  self.primary_key = 'id'
  
  attr_accessor :name, :email
  
  def initialize(name, email)
    @name = name
    @email = email
  end
  
  def save
    # This will trigger the prepended module's save method
    puts "Saving user: #{@name}"
    
    # Notify observers of the change
    notify_change(:saved, self)
    
    true
  end
  
  def expensive_operation
    # This would trigger autoload of ExpensiveModule if we used it
    # ExpensiveModule.process(self)
    
    # Instead, just do something simple
    sleep(0.1)
    "Done"
  end
end

# Observer for the User model
class UserObserver
  def update(attribute, user)
    puts "User #{attribute}: #{user.name}"
  end
end

# Dynamic module loading based on configuration
module DynamicLoader
  def self.load_plugins(plugin_names)
    plugin_names.each do |plugin_name|
      begin
        require "plugins/#{plugin_name}"
        puts "Loaded plugin: #{plugin_name}"
      rescue LoadError
        puts "Plugin not found: #{plugin_name}"
      end
    end
  end
end

# Using the classes
user = User.new("Alice", "alice@example.com")
observer = UserObserver.new
user.add_observer(observer)

# Benchmark a method call
user.benchmark_method(:expensive_operation)

# Save triggers prepended module and observers
user.save

# Try to load some plugins dynamically
DynamicLoader.load_plugins(['analytics', 'metrics', 'reporting'])

# Using Set in a method
def unique_items(array)
  # Actually, let's use Set
  require 'set'
  Set.new(array).to_a
end

puts "Unique items: #{unique_items([1, 2, 2, 3, 3, 3])}"