# Test Pattern 4: Conditional require and require_relative
# Tests requires within conditional blocks and dynamic loading

require 'yaml'
require 'json'
require_relative 'config/settings'
require_relative 'config/database'

# Platform-specific requires
if RUBY_PLATFORM =~ /darwin/
  require 'osx/cocoa'
elsif RUBY_PLATFORM =~ /linux/
  require 'gtk3'
elsif RUBY_PLATFORM =~ /mingw|mswin/
  require 'win32ole'
end

# Version-specific requires
if RUBY_VERSION >= '3.0.0'
  require 'fiber/scheduler'
  require 'ractor'
else
  require 'thread'
  require 'monitor'
end

# Environment-based requires
if ENV['RAILS_ENV'] == 'test'
  require 'rspec'
  require 'factory_bot'
  require 'database_cleaner'
elsif ENV['RAILS_ENV'] == 'development'
  require 'pry'
  require 'better_errors'
  require 'binding_of_caller'
end

# Optional requires with error handling
begin
  require 'redis'
  REDIS_AVAILABLE = true
rescue LoadError
  REDIS_AVAILABLE = false
end

begin
  require 'memcached'
  MEMCACHED_AVAILABLE = true
rescue LoadError
  MEMCACHED_AVAILABLE = false
end

# Not using: database, gtk3, win32ole, fiber/scheduler, ractor, thread, monitor,
# rspec, factory_bot, database_cleaner, better_errors, binding_of_caller, memcached

class ConfigManager
  def initialize
    # Using YAML
    @config = YAML.load_file('config.yml') rescue {}

    # Using settings from require_relative
    @settings = Settings.load

    # Platform-specific code
    setup_platform_specific

    # Setup cache if available
    setup_cache
  end

  def to_json
    # Using JSON
    JSON.pretty_generate({
      config: @config,
      settings: @settings.to_h,
      platform: RUBY_PLATFORM,
      cache: @cache_type
    })
  end

  private

  def setup_platform_specific
    if RUBY_PLATFORM =~ /darwin/
      # Would use osx/cocoa here if we were on Mac
      puts "Running on macOS"
      # OSX::NSApplication.sharedApplication if this were real
    end

    # For Linux/Windows, the requires would be used here
  end

  def setup_cache
    if REDIS_AVAILABLE
      # Using Redis if available
      require 'redis'  # Requiring again to be sure
      @cache = Redis.new
      @cache_type = 'redis'
    elsif MEMCACHED_AVAILABLE
      # Would use memcached here
      @cache_type = 'none'  # Not actually using it
    else
      @cache_type = 'none'
    end
  end

  def development_helpers
    if ENV['RAILS_ENV'] == 'development'
      # Using pry in development
      require 'pry'
      binding.pry if ENV['DEBUG']
    end
  end
end

# Version-specific code
if RUBY_VERSION >= '3.0.0'
  # Would use Ractor or Fiber scheduler here in Ruby 3+
  puts "Ruby 3+ features available"
else
  # Would use Thread/Monitor in older Ruby
  puts "Using compatibility mode"
end

manager = ConfigManager.new
puts manager.to_json