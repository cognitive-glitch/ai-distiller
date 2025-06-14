# frozen_string_literal: true

require 'json'
require 'time'

# Module demonstrating mixins, include/extend, blocks and basic metaprogramming
module Loggable
  # Method added to instances when included
  def log(message, level: :info)
    timestamp = Time.now.iso8601
    puts "[#{timestamp}] #{level.upcase}: #{message}"
  end
  
  # Module method for class-level functionality
  def self.included(base)
    base.extend(ClassMethods)
    base.class_eval do
      attr_accessor :logger_enabled
    end
  end
  
  # Methods added to class when module is included
  module ClassMethods
    def enable_logging
      define_method(:log_enabled?) { @logger_enabled || false }
    end
  end
end

# Module for JSON serialization mixin
module Serializable
  def to_json(*args)
    JSON.generate(serializable_hash, *args)
  end
  
  def from_json(json_string)
    data = JSON.parse(json_string)
    load_from_hash(data)
  end
  
  private
  
  # Abstract method to be implemented by including classes
  def serializable_hash
    raise NotImplementedError, "Must implement serializable_hash"
  end
  
  def load_from_hash(hash)
    raise NotImplementedError, "Must implement load_from_hash"
  end
end

# Enumerable module usage with blocks and yield
module RoleManager
  VALID_ROLES = %w[user admin moderator guest].freeze
  
  def self.each_role
    return enum_for(:each_role) unless block_given?
    
    VALID_ROLES.each do |role|
      yield role, role.upcase
    end
  end
  
  def self.find_role(pattern)
    VALID_ROLES.find { |role| role.match?(pattern) }
  end
  
  # Demonstrate various block syntaxes
  def self.role_stats
    roles_with_lengths = VALID_ROLES.map { |role| [role, role.length] }
    roles_with_lengths.select { |_, length| length > 4 }
  end
end

# Class demonstrating mixins and blocks
class Document
  include Loggable
  include Serializable
  
  attr_reader :title, :content, :author, :created_at
  
  def initialize(title, content, author)
    @title = title
    @content = content
    @author = author
    @created_at = Time.now
    @logger_enabled = true
  end
  
  # Method using blocks with yield
  def process_content
    return enum_for(:process_content) unless block_given?
    
    log("Processing content for: #{@title}")
    
    lines = @content.split("\n")
    lines.each_with_index do |line, index|
      processed_line = yield(line, index)
      lines[index] = processed_line if processed_line
    end
    
    @content = lines.join("\n")
  end
  
  # Method with different block patterns
  def find_sections(pattern = /^#/)
    sections = []
    @content.each_line.with_index do |line, index|
      sections << { line: index + 1, content: line.strip } if line.match?(pattern)
    end
    
    if block_given?
      sections.each { |section| yield section }
    else
      sections
    end
  end
  
  # Basic method_missing metaprogramming
  def method_missing(method_name, *args, &block)
    if method_name.to_s.start_with?('find_by_')
      attribute = method_name.to_s.sub('find_by_', '')
      search_value = args.first
      
      case attribute
      when 'author'
        @author == search_value
      when 'title'
        @title.include?(search_value)
      else
        super
      end
    else
      super
    end
  end
  
  def respond_to_missing?(method_name, include_private = false)
    method_name.to_s.start_with?('find_by_') || super
  end
  
  private
  
  def serializable_hash
    {
      title: @title,
      content: @content,
      author: @author,
      created_at: @created_at.iso8601
    }
  end
  
  def load_from_hash(hash)
    @title = hash['title']
    @content = hash['content']
    @author = hash['author']
    @created_at = Time.parse(hash['created_at']) if hash['created_at']
  end
end

# Class with extend usage (adds module methods as class methods)
class DocumentRepository
  extend RoleManager
  
  def initialize
    @documents = []
  end
  
  def add_document(document)
    @documents << document
    log_action("Added document: #{document.title}")
  end
  
  # Method using blocks for filtering
  def filter_documents(&block)
    return @documents.dup unless block_given?
    @documents.select(&block)
  end
  
  # Method with proc/lambda demonstration
  def sort_documents(sort_proc = nil)
    sort_logic = sort_proc || lambda { |a, b| a.created_at <=> b.created_at }
    @documents.sort(&sort_logic)
  end
  
  protected
  
  def log_action(message)
    puts "[DocumentRepository] #{message}"
  end
  
  private
  
  def validate_document(document)
    document.is_a?(Document) && !document.title.empty?
  end
end