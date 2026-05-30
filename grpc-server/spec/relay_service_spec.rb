# grpc-server/spec/relay_service_spec.rb
require 'spec_helper'

RSpec.describe BattleshipRelayService do
  let(:service) { described_class.new }

  describe '#create_game' do
    it 'returns a 6-character alphanumeric game_id' do
      req  = Battleship::CreateGameRequest.new
      resp = service.create_game(req, nil)
      expect(resp.game_id).to match(/\A[A-Z0-9]{6}\z/)
    end

    it 'registers the game in the sessions map' do
      req  = Battleship::CreateGameRequest.new
      resp = service.create_game(req, nil)
      expect(service.sessions).to have_key(resp.game_id)
    end
  end

  describe '#join_game' do
    it 'returns success: true for a valid game_id' do
      create_resp = service.create_game(Battleship::CreateGameRequest.new, nil)
      req  = Battleship::JoinGameRequest.new(game_id: create_resp.game_id)
      resp = service.join_game(req, nil)
      expect(resp.success).to be true
      expect(resp.error_message).to eq('')
    end

    it 'returns success: false for an unknown game_id' do
      req  = Battleship::JoinGameRequest.new(game_id: 'NOPE99')
      resp = service.join_game(req, nil)
      expect(resp.success).to be false
      expect(resp.error_message).to include('not found')
    end

    it 'returns success: false when a second joiner tries to join' do
      create_resp = service.create_game(Battleship::CreateGameRequest.new, nil)
      service.join_game(Battleship::JoinGameRequest.new(game_id: create_resp.game_id), nil)
      resp = service.join_game(Battleship::JoinGameRequest.new(game_id: create_resp.game_id), nil)
      expect(resp.success).to be false
      expect(resp.error_message).to include('full')
    end
  end
end
