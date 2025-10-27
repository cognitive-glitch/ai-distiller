# Test Pattern 3: Rails-specific requires and complex patterns
# Tests Rails components and nested requires

require 'rails/all'
require 'active_model/railtie'
require 'active_job/railtie'
require 'action_cable/engine'
require 'active_storage/engine'
require 'action_mailer/railtie'

# Gems commonly used with Rails
require 'devise'
require 'cancancan'
require 'paperclip'
require 'sidekiq'

# Not using: active_model/railtie, active_job/railtie, action_cable/engine,
# active_storage/engine, action_mailer/railtie, devise, cancancan, paperclip

module ApplicationHelper
  # Helper methods would go here
end

class ApplicationController < ActionController::Base
  # This implicitly uses rails/all
  protect_from_forgery with: :exception

  before_action :configure_permitted_parameters, if: :devise_controller?

  protected

  def configure_permitted_parameters
    # This method suggests Devise is used, but it's not actually called
    # in this example, so Devise import might be marked as unused
  end
end

class JobsController < ApplicationController
  def index
    # Using Rails features (from rails/all)
    @jobs = Job.all

    # Using Sidekiq for background job
    ProcessingJob.perform_later(@jobs.pluck(:id))

    respond_to do |format|
      format.html
      format.json { render json: @jobs }
    end
  end
end

class ProcessingJob < ApplicationJob
  # This uses ActiveJob from rails/all
  queue_as :default

  # Using Sidekiq as the adapter
  self.queue_adapter = :sidekiq

  def perform(job_ids)
    job_ids.each do |id|
      # Process each job
      Rails.logger.info "Processing job #{id}"
    end
  end
end

# Model using Rails
class Job < ApplicationRecord
  # These would typically use other gems like:
  # has_attached_file :document (Paperclip)
  # But we're not using them in this example

  validates :title, presence: true
  validates :description, length: { minimum: 10 }

  scope :active, -> { where(active: true) }
  scope :recent, -> { order(created_at: :desc) }
end

# Configuration that proves Rails is used
Rails.application.configure do
  config.eager_load = false
  config.consider_all_requests_local = true
end