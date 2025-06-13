# Basic Ruby example for testing

require 'json'
require 'net/http'
require_relative 'helpers/string_utils'

# A person class demonstrating Ruby features
class Person < ActiveRecord::Base
  include Comparable
  extend Forwardable
  
  # Constants
  MAX_AGE = 150
  DEFAULT_COUNTRY = "USA"
  
  # Class variables
  @@count = 0
  
  # Instance variables with attr_accessor
  attr_accessor :name, :email
  attr_reader :id, :created_at
  attr_writer :password
  
  # Initialize method (constructor)
  def initialize(name, age = 18)
    @name = name
    @age = age
    @id = SecureRandom.uuid
    @created_at = Time.now
    @@count += 1
  end
  
  # Public instance method
  def full_name
    "#{first_name} #{last_name}"
  end
  
  # Method with various parameter types
  def complex_method(required, optional = "default", *args, keyword:, optional_keyword: nil, **kwargs, &block)
    # Method implementation
    yield if block_given?
  end
  
  # Private methods
  private
  
  def validate_age
    raise ArgumentError, "Invalid age" unless @age.between?(0, MAX_AGE)
  end
  
  # Protected method
  protected
  
  def internal_id
    "PERSON-#{@id}"
  end
  
  # Class methods
  class << self
    def count
      @@count
    end
    
    def find_by_name(name)
      # Database lookup simulation
      where(name: name).first
    end
  end
  
  # Alternative class method definition
  def self.create_guest
    new("Guest User", 0)
  end
  
  # Method aliasing
  alias_method :display_name, :full_name
  
  # Dynamic method definition
  define_method :dynamic_greeting do |greeting = "Hello"|
    "#{greeting}, #{@name}!"
  end
  
  # Operator overloading
  def <=>(other)
    @age <=> other.age
  end
  
  # Method missing for dynamic attributes
  def method_missing(method_name, *args, &block)
    if method_name.to_s.start_with?("find_by_")
      # Dynamic finder implementation
      super
    else
      super
    end
  end
end

# Module definition
module Validatable
  # Module constants
  VALIDATION_RULES = {
    email: /\A[\w+\-.]+@[a-z\d\-]+(\.[a-z\d\-]+)*\.[a-z]+\z/i,
    phone: /\A\d{10}\z/
  }
  
  def self.included(base)
    base.extend(ClassMethods)
  end
  
  module ClassMethods
    def validates(attribute, options = {})
      # Validation DSL
    end
  end
  
  # Instance methods
  def valid?
    validate
    errors.empty?
  end
  
  private
  
  def validate
    # Validation logic
  end
end

# Singleton class example
class Configuration
  include Singleton
  
  attr_accessor :debug_mode, :api_key
  
  def initialize
    @debug_mode = false
    @api_key = ENV['API_KEY']
  end
end

# Struct example
User = Struct.new(:username, :email, :role) do
  def admin?
    role == 'admin'
  end
end

# Module with enumerable
module Statistics
  extend Enumerable
  
  def self.each(&block)
    data.each(&block)
  end
  
  private
  
  def self.data
    @data ||= []
  end
end

# Global method
def global_helper(input)
  input.to_s.upcase
end

# Lambda and Proc examples
uppercase = ->(str) { str.upcase }
multiplier = Proc.new { |x, y| x * y }

# Constant at top level
API_VERSION = "1.0.0"

# Conditional class definition
if defined?(Rails)
  class RailsSpecificClass
    # Rails-specific implementation
  end
end

# Exception class
class CustomError < StandardError
  attr_reader :code
  
  def initialize(message, code = nil)
    super(message)
    @code = code
  end
end

# Namespace example
module MyApp
  module Models
    class User < Person
      # Nested class
    end
  end
  
  # Module function
  module_function
  
  def version
    "1.0.0"
  end
end

# Method with block
def process_items(items)
  items.each do |item|
    yield item if block_given?
  end
end

# Define method on specific object (singleton method)
special_object = Object.new
def special_object.special_method
  "I'm special!"
end