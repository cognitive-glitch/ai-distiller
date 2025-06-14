# frozen_string_literal: true

# A collection of simple utilities for user management
# Demonstrates basic Ruby features: classes, modules, constants, visibility

# Module containing utility functions and constants
module UserUtils
  # Public constant for minimum password length
  MIN_PASSWORD_LENGTH = 8
  
  # Public constant with email regex
  EMAIL_REGEX = /\A[\w+\-.]+@[a-z\d\-]+(\.[a-z\d\-]+)*\.[a-z]+\z/i.freeze
  
  # Public module method for email validation
  def self.valid_email?(email)
    return false unless email.is_a?(String)
    email.match?(EMAIL_REGEX)
  end
  
  # Public module method for password validation
  def self.strong_password?(password)
    return false unless password.is_a?(String)
    return false if password.length < MIN_PASSWORD_LENGTH
    
    has_letter = password.match?(/[a-zA-Z]/)
    has_digit = password.match?(/\d/)
    has_letter && has_digit
  end
  
  private
  
  # Private module method (not commonly used but valid Ruby)
  def self.internal_helper
    "This is a private module method"
  end
end

# Basic class demonstrating Ruby conventions
class User
  # Use attr_accessor for public read/write access
  attr_accessor :name, :email
  
  # Use attr_reader for public read-only access
  attr_reader :id, :created_at
  
  # Class variable (shared across all instances)
  @@user_count = 0
  
  # Public class method to get user count
  def self.count
    @@user_count
  end
  
  # Constructor method
  def initialize(name, email)
    @id = generate_id
    @name = name
    @email = email
    @created_at = Time.now
    @@user_count += 1
  end
  
  # Public instance method
  def to_s
    "User(#{@id}): #{@name} <#{@email}>"
  end
  
  # Public instance method with validation
  def valid?
    UserUtils.valid_email?(@email) && !@name.empty?
  end
  
  # Protected method (can be called by other instances of same class)
  protected
  
  def compare_creation_time(other_user)
    @created_at <=> other_user.created_at
  end
  
  # Private methods (only callable from within the same instance)
  private
  
  def generate_id
    "user_#{Time.now.to_i}_#{rand(1000)}"
  end
end

# Subclass demonstrating inheritance
class AdminUser < User
  attr_reader :permissions
  
  def initialize(name, email, permissions = [])
    super(name, email)
    @permissions = permissions || []
  end
  
  # Override parent method
  def to_s
    "#{super} [Admin]"
  end
  
  # Public method specific to AdminUser
  def has_permission?(permission)
    @permissions.include?(permission)
  end
  
  private
  
  # Private method in subclass
  def validate_permissions
    @permissions.all? { |p| p.is_a?(String) }
  end
end