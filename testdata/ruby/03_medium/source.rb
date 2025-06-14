# frozen_string_literal: true

require 'forwardable'

# Advanced metaprogramming with define_method, class_eval, and hooks
module MetaProgrammingUtils
  # Hook called when module is included
  def self.included(base)
    base.extend(ClassMethods)
    base.class_eval do
      # Dynamically add instance variables and accessors
      attr_accessor :metadata
      
      # Add a class instance variable
      @dynamic_methods = []
    end
  end
  
  module ClassMethods
    # Dynamically define getter methods
    def add_dynamic_accessor(name, default_value = nil)
      @dynamic_methods ||= []
      @dynamic_methods << name
      
      # Define getter
      define_method(name) do
        instance_variable_get("@#{name}") || default_value
      end
      
      # Define setter
      define_method("#{name}=") do |value|
        instance_variable_set("@#{name}", value)
      end
      
      # Define query method
      define_method("#{name}?") do
        !!instance_variable_get("@#{name}")
      end
    end
    
    # Create methods that delegate to another object
    def delegate_to(target, *methods)
      methods.each do |method|
        define_method(method) do |*args, &block|
          target_obj = instance_variable_get("@#{target}")
          target_obj.send(method, *args, &block)
        end
      end
    end
    
    # Class method to track dynamic methods
    def dynamic_methods
      @dynamic_methods || []
    end
  end
end

# Module demonstrating eigenclass manipulation
module EigenclassDemo
  def self.extended(base)
    # Add methods to the eigenclass (singleton methods)
    base.instance_eval do
      def singleton_method_added(method_name)
        puts "Singleton method '#{method_name}' added to #{self}"
        super
      end
    end
  end
  
  # This becomes a class method when extended
  def custom_new(*args, &block)
    instance = allocate
    instance.send(:initialize, *args, &block) if instance.respond_to?(:initialize, true)
    
    # Add a singleton method to this specific instance
    instance.define_singleton_method(:created_with_custom_new?) { true }
    
    instance
  end
end

# Advanced module with hooks and callbacks
module Trackable
  def self.included(base)
    base.extend(ClassMethods)
    base.class_eval do
      @tracked_methods = []
    end
  end
  
  module ClassMethods
    # Method to track calls to specific methods
    def track_method(method_name)
      @tracked_methods ||= []
      @tracked_methods << method_name
      
      original_method = instance_method(method_name)
      
      define_method(method_name) do |*args, &block|
        track_method_call(method_name, args)
        original_method.bind(self).call(*args, &block)
      end
    end
    
    def tracked_methods
      @tracked_methods || []
    end
  end
  
  private
  
  def track_method_call(method_name, args)
    @method_calls ||= []
    @method_calls << {
      method: method_name,
      args: args,
      timestamp: Time.now
    }
  end
  
  protected
  
  def method_call_history
    @method_calls || []
  end
end

# Class demonstrating advanced metaprogramming
class ConfigurableModel
  include MetaProgrammingUtils
  include Trackable
  extend EigenclassDemo
  extend Forwardable
  
  # Use forwardable gem for delegation
  def_delegators :@config, :[], :[]=, :keys, :values
  
  def initialize(initial_config = {})
    @config = initial_config
    @metadata = {}
    
    # Dynamically define methods based on config keys
    initial_config.each do |key, value|
      self.class.add_dynamic_accessor(key, value)
      instance_variable_set("@#{key}", value)
    end
  end
  
  # Method using class_eval for dynamic method definition
  def self.add_validation(field, &validation_block)
    class_eval do
      define_method("validate_#{field}") do
        value = send(field)
        validation_block.call(value)
      end
      
      alias_method "#{field}_valid?", "validate_#{field}"
    end
  end
  
  # Method demonstrating eval usage (carefully controlled)
  def evaluate_expression(expression, context = {})
    # Create a clean binding with only allowed variables
    binding_context = binding
    context.each do |key, value|
      binding_context.local_variable_set(key, value)
    end
    
    # Only allow simple mathematical expressions
    if expression.match?(/\A[\d\s+\-*\/().]+\z/)
      binding_context.eval(expression)
    else
      raise ArgumentError, "Invalid expression"
    end
  end
  
  # Method using define_singleton_method
  def add_instance_method(method_name, &block)
    define_singleton_method(method_name, &block)
  end
  
  # Hook method called when methods are added
  def self.method_added(method_name)
    puts "Method '#{method_name}' added to class #{self}"
    super
  end
  
  private
  
  # Private method using send and respond_to?
  def invoke_if_exists(method_name, *args)
    if respond_to?(method_name, true)
      send(method_name, *args)
    else
      nil
    end
  end
  
  # Method using instance_eval
  def configure(&block)
    instance_eval(&block) if block_given?
  end
end

# Class demonstrating const_missing hook
class DynamicConstants
  # Hook called when a constant is missing
  def self.const_missing(const_name)
    if const_name.to_s.start_with?('DYNAMIC_')
      # Dynamically create constants
      const_value = const_name.to_s.gsub('DYNAMIC_', '').downcase
      const_set(const_name, const_value)
    else
      super
    end
  end
  
  # Class method that uses const_get and const_set
  def self.create_constant(name, value)
    const_set(name.to_s.upcase, value)
  end
  
  def self.list_constants
    constants.map { |const| [const, const_get(const)] }.to_h
  end
end

# Example usage class
class SmartDocument < ConfigurableModel
  track_method :content=
  
  add_dynamic_accessor :title, "Untitled"
  add_dynamic_accessor :author, "Anonymous"
  add_dynamic_accessor :content, ""
  
  # Add validation using the class method
  add_validation(:title) { |title| !title.empty? }
  add_validation(:content) { |content| content.length > 10 }
  
  def initialize(config = {})
    super(config)
    configure_smart_features
  end
  
  private
  
  def configure_smart_features
    # Use instance_eval to add methods dynamically
    configure do
      def word_count
        content.split.length
      end
      
      def summary(length = 50)
        content[0, length] + (content.length > length ? "..." : "")
      end
    end
  end
end