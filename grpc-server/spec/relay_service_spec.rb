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

  describe '#game_stream' do
    it 'relays a message from player A to player B' do
      create_resp = service.create_game(Battleship::CreateGameRequest.new, nil)
      game_id     = create_resp.game_id

      msg_a = Battleship::GameMessage.new(game_id: game_id, command: 'SHOT', args: { 'coord' => 'A1' })
      msg_b = Battleship::GameMessage.new(game_id: game_id, command: 'SHOT_RESULT', args: { 'result' => 'hit' })

      a_request_queue = Queue.new
      b_request_queue = Queue.new
      a_request_queue << msg_a

      a_requests = Enumerator.new { |y| loop { item = a_request_queue.pop; break if item == :done; y << item } }
      b_requests = Enumerator.new { |y| loop { item = b_request_queue.pop; break if item == :done; y << item } }

      a_enum = service.game_stream(a_requests, nil)
      b_enum = service.game_stream(b_requests, nil)

      sleep 0.1

      b_request_queue << msg_b
      sleep 0.1

      receiver_thread = Thread.new { a_enum.next rescue nil }
      received = receiver_thread.join(2)&.value  # 2-second timeout; nil if hung

      expect(received).not_to be_nil
      expect(received.command).to eq('SHOT_RESULT')

      a_request_queue << :done
      b_request_queue << :done
    end

    it 'rejects a stream with an unknown game_id by sending eof' do
      bad_msg = Battleship::GameMessage.new(game_id: 'XXXXXX', command: 'SHOT')
      req_q   = Queue.new
      req_q   << bad_msg
      req_q   << :done

      requests = Enumerator.new { |y| loop { item = req_q.pop; break if item == :done; y << item } }
      enum     = service.game_stream(requests, nil)

      sleep 0.1
      expect { enum.next }.to raise_error(StopIteration)
    end
  end
end
