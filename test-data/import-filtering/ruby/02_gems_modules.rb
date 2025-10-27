# Test Pattern 2: Gem require and module inclusion
# Tests loading gems and including/extending modules

require 'active_support/all'
require 'active_record'
require 'logger'
require 'singleton'
require 'forwardable'

# Custom modules for testing
module Validatable
  def valid?
    validate
  end

  def validate
    # Validation logic would go here
    true
  end
end

module Trackable
  def self.included(base)
    base.extend(ClassMethods)
  end

  module ClassMethods
    def track_method(method_name)
      # Tracking logic
    end
  end

  def track_changes
    @changes ||= {}
  end
end

module Serializable
  def to_json
    # Custom JSON serialization
  end
end

# Not using: active_record, singleton, forwardable, Serializable

class Application
  include Singleton  # Wait, this is actually used!
  include Trackable
  extend Forwardable  # And this too!

  # Using Logger
  def initialize
    @logger = Logger.new(STDOUT)
    @logger.info("Application initialized")
  end

  # Using forwardable to delegate
  def_delegators :@logger, :info, :warn, :error

  # Using ActiveSupport
  def process_data(data)
    # Using ActiveSupport's blank? method
    return nil if data.blank?

    # Using ActiveSupport's pluralize
    word = "person"
    info "Processing #{word.pluralize}"

    # Using track_changes from Trackable
    track_changes[:processed] = true

    # Using ActiveSupport's days method
    deadline = 3.days.from_now
    info "Deadline: #{deadline}"

    data
  end
end

class User
  include Validatable
  # Not including Serializable even though it's defined

  attr_accessor :name, :email

  def initialize(name, email)
    @name = name
    @email = email
  end

  def validate
    !@name.nil? && !@email.nil?
  end
end

# Actually, let me correct - Singleton and Forwardable ARE used
# So not using: active_record, Serializable

app = Application.instance
app.process_data("test data")

user = User.new("John", "john@example.com")
puts "User valid? #{user.valid?}"