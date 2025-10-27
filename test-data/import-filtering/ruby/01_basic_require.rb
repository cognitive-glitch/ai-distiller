# Test Pattern 1: Basic require and include
# Tests simple file/gem loading and module inclusion

require 'json'
require 'yaml'
require 'net/http'
require 'uri'
require 'fileutils'
require_relative 'helpers/string_utils'
require_relative 'helpers/array_utils'

# Module to include
module Enumerable
  # Adding custom method to Enumerable
  def average
    sum.to_f / size
  end
end

# Not using: yaml, uri, fileutils, array_utils

class DataProcessor
  include Enumerable  # Actually this doesn't make sense for a class, but testing

  def initialize
    @data = []
  end

  def process(data_string)
    # Using JSON
    parsed = JSON.parse(data_string)

    # Using Net::HTTP
    response = fetch_remote_data(parsed['url']) if parsed['url']

    # Using string_utils (from require_relative)
    cleaned = StringUtils.clean(parsed['content'] || '')

    {
      parsed: parsed,
      cleaned: cleaned,
      response: response
    }
  end

  private

  def fetch_remote_data(url_string)
    # Using Net::HTTP but not URI directly (it's used internally by Net::HTTP)
    response = Net::HTTP.get_response(URI(url_string))
    response.body if response.is_a?(Net::HTTPSuccess)
  rescue => e
    puts "Error fetching data: #{e.message}"
    nil
  end
end

# Using the processor
processor = DataProcessor.new
result = processor.process('{"url": "http://example.com", "content": "  Hello  "}')
puts result.inspect