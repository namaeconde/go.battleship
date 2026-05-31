# grpc-server/spec/spec_helper.rb
$LOAD_PATH.unshift File.join(__dir__, '..', 'lib')

require 'grpc'
require 'battleship_pb'
require 'battleship_services_pb'
require_relative '../server'
