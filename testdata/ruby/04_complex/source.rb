# frozen_string_literal: true

require 'ostruct'

# Complex DSL creation, method chaining, advanced metaprogramming
module DSLBuilder
  def self.included(base)
    base.extend(ClassMethods)
  end

  module ClassMethods
    # Create a DSL method that accepts a block
    def dsl_method(name, &default_block)
      define_method(name) do |*args, &block|
        # Create a new DSL context
        dsl_context = DSLContext.new(self)

        # Execute default block first if provided
        dsl_context.instance_eval(&default_block) if default_block

        # Then execute the user-provided block
        dsl_context.instance_eval(&block) if block

        # Return self for chaining
        self
      end
    end

    # Create chainable setter methods
    def chainable_attr(*names)
      names.each do |name|
        define_method(name) do |value = nil|
          if value.nil?
            instance_variable_get("@#{name}")
          else
            instance_variable_set("@#{name}", value)
            self # Return self for chaining
          end
        end
      end
    end
  end

  # Internal DSL context class
  class DSLContext
    def initialize(target)
      @target = target
    end

    # Forward missing methods to target
    def method_missing(method_name, *args, &block)
      if @target.respond_to?(method_name, true)
        @target.send(method_name, *args, &block)
      else
        super
      end
    end

    def respond_to_missing?(method_name, include_private = false)
      @target.respond_to?(method_name, include_private) || super
    end
  end
end

# Module for creating fluent interfaces
module FluentInterface
  def self.included(base)
    base.extend(ClassMethods)
  end

  module ClassMethods
    # Create a fluent builder pattern
    def fluent_builder(*method_names)
      method_names.each do |method_name|
        define_method(method_name) do |value = nil, &block|
          if block_given?
            # If block is given, create a sub-builder
            sub_builder = self.class.new
            sub_builder.instance_eval(&block)
            instance_variable_set("@#{method_name}", sub_builder)
          elsif value
            instance_variable_set("@#{method_name}", value)
          end

          self # Always return self for chaining
        end

        # Create getter method
        define_method("get_#{method_name}") do
          instance_variable_get("@#{method_name}")
        end
      end
    end
  end
end

# Advanced reflection and dynamic class creation
module ReflectionUtils
  # Dynamically create classes at runtime
  def self.create_class(class_name, parent_class = Object, &block)
    new_class = Class.new(parent_class) do
      define_method(:initialize) do |*args|
        @attributes = {}
        args.each_with_index do |arg, index|
          @attributes["attr_#{index}".to_sym] = arg
        end
        super(*args) if defined?(super)
      end

      # Add attribute accessors
      define_method(:get_attribute) do |key|
        @attributes[key.to_sym]
      end

      define_method(:set_attribute) do |key, value|
        @attributes[key.to_sym] = value
        self
      end

      # Allow custom class definition
      class_eval(&block) if block_given?
    end

    # Set the class name in the constant table
    Object.const_set(class_name.to_sym, new_class) unless Object.const_defined?(class_name.to_sym)
    new_class
  end

  # Analyze class structure
  def self.analyze_class(klass)
    {
      name: klass.name,
      superclass: klass.superclass&.name,
      included_modules: klass.included_modules.map(&:name),
      instance_methods: klass.instance_methods(false),
      private_methods: klass.private_instance_methods(false),
      protected_methods: klass.protected_instance_methods(false),
      constants: klass.constants
    }
  end
end

# Complex DSL for building configurations
class ConfigurationBuilder
  include DSLBuilder
  include FluentInterface

  chainable_attr :name, :version, :description
  fluent_builder :database, :cache, :logging

  def initialize
    @settings = {}
    @nested_configs = {}
  end

  # DSL method for defining settings
  dsl_method :configure do
    def setting(key, value = nil, &block)
      if block_given?
        # Nested configuration
        nested_builder = ConfigurationBuilder.new
        nested_builder.instance_eval(&block)
        @target.instance_variable_get(:@nested_configs)[key] = nested_builder
      else
        @target.instance_variable_get(:@settings)[key] = value
      end
    end

    def environment(env_name, &block)
      setting("environment_#{env_name}", &block)
    end
  end

  # Method chaining for array operations
  def add_middleware(middleware_class, *options)
    @middleware ||= []
    @middleware << { class: middleware_class, options: options }
    self
  end

  def add_plugin(plugin_name, &configuration_block)
    @plugins ||= {}

    if configuration_block
      plugin_config = PluginConfiguration.new
      plugin_config.instance_eval(&configuration_block)
      @plugins[plugin_name] = plugin_config
    else
      @plugins[plugin_name] = true
    end

    self
  end

  # Dynamic method creation based on patterns
  def method_missing(method_name, *args, &block)
    method_str = method_name.to_s

    case method_str
    when /^with_(.+)$/
      # with_* methods for fluent configuration
      attribute_name = $1
      instance_variable_set("@#{attribute_name}", args.first || true)
      self
    when /^enable_(.+)$/
      # enable_* methods
      feature_name = $1
      @enabled_features ||= []
      @enabled_features << feature_name
      self
    when /^(.+)_callback$/
      # *_callback methods
      callback_name = $1
      @callbacks ||= {}
      @callbacks[callback_name] = block
      self
    else
      super
    end
  end

  def respond_to_missing?(method_name, include_private = false)
    method_str = method_name.to_s
    method_str.match?(/^(with_|enable_|.+_callback$)/) || super
  end

  # Convert configuration to hash
  def to_hash
    result = {
      settings: @settings,
      nested_configs: @nested_configs.transform_values(&:to_hash),
      middleware: @middleware,
      plugins: @plugins,
      enabled_features: @enabled_features,
      callbacks: @callbacks&.keys
    }

    # Add chainable attributes
    %w[name version description].each do |attr|
      value = instance_variable_get("@#{attr}")
      result[attr.to_sym] = value if value
    end

    result.compact
  end

  private

  def validate_configuration
    errors = []
    errors << "Name is required" unless @name
    errors << "Version is required" unless @version
    errors
  end
end

# Plugin configuration DSL
class PluginConfiguration
  def initialize
    @options = {}
  end

  def option(key, value)
    @options[key] = value
    self
  end

  def timeout(seconds)
    option(:timeout, seconds)
  end

  def retries(count)
    option(:retries, count)
  end

  def method_missing(method_name, *args)
    if args.length == 1
      option(method_name, args.first)
    else
      super
    end
  end

  def respond_to_missing?(method_name, include_private = false)
    true # Accept any method as an option
  end

  def to_hash
    @options
  end
end

# Factory for creating dynamic classes with DSL
class DynamicClassFactory
  def self.create_model(name, &definition)
    model_class = ReflectionUtils.create_class(name) do
      include DSLBuilder

      attr_reader :attributes

      def initialize
        @attributes = {}
      end

      # Dynamic attribute methods
      def self.attr_with_validation(name, &validator)
        define_method(name) do
          @attributes[name]
        end

        define_method("#{name}=") do |value|
          if validator
            unless validator.call(value)
              raise ArgumentError, "Invalid value for #{name}: #{value}"
            end
          end
          @attributes[name] = value
        end
      end

      # Class-level DSL for field definitions
      def self.field(name, type = :string, **options)
        attr_with_validation(name) do |value|
          case type
          when :string
            value.is_a?(String)
          when :integer
            value.is_a?(Integer)
          when :boolean
            [true, false].include?(value)
          else
            true
          end
        end

        # Add to schema
        @schema ||= {}
        @schema[name] = { type: type, options: options }
      end

      def self.schema
        @schema || {}
      end
    end

    # Execute the definition block in the class context
    model_class.class_eval(&definition) if definition

    model_class
  end
end