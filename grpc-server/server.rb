# grpc-server/server.rb
$LOAD_PATH.unshift File.join(__dir__, 'lib')

require 'grpc'
require 'battleship_pb'
require 'battleship_services_pb'

class BattleshipRelayService < Battleship::BattleshipRelay::Service
  attr_reader :sessions

  def initialize
    @sessions = {}
    @mutex    = Mutex.new
  end

  def create_game(_req, _call)
    game_id = generate_game_id
    @mutex.synchronize { @sessions[game_id] = { player_b_joined: false, streams: [] } }
    Battleship::CreateGameResponse.new(game_id: game_id)
  end

  def join_game(req, _call)
    game_id = req.game_id
    @mutex.synchronize do
      session = @sessions[game_id]
      return Battleship::JoinGameResponse.new(success: false, error_message: "Game #{game_id} not found") unless session
      return Battleship::JoinGameResponse.new(success: false, error_message: "Game #{game_id} is full") if session[:player_b_joined]
      session[:player_b_joined] = true
    end
    Battleship::JoinGameResponse.new(success: true)
  end

  def game_stream(requests)
    game_id = nil
    queue   = Queue.new

    requests.each do |msg|
      if game_id.nil?
        game_id = msg.game_id
        register_stream(game_id, queue)
      end
      relay_to_opponent(game_id, queue, msg)
    end

    Enumerator.new do |y|
      loop { y << queue.pop }
    end
  end

  private

  def generate_game_id
    chars = ('A'..'Z').to_a + ('0'..'9').to_a
    6.times.map { chars.sample }.join
  end

  def register_stream(game_id, queue)
    @mutex.synchronize do
      @sessions[game_id] ||= { player_b_joined: false, streams: [] }
      @sessions[game_id][:streams] << queue
    end
  end

  def relay_to_opponent(game_id, own_queue, msg)
    @mutex.synchronize do
      session = @sessions[game_id]
      return unless session
      session[:streams].each do |q|
        q << msg unless q.equal?(own_queue)
      end
    end
  end
end

if __FILE__ == $PROGRAM_NAME
  port   = ENV.fetch('PORT', '8080')
  server = GRPC::RpcServer.new
  server.add_http2_port("0.0.0.0:#{port}", :this_port_is_insecure)
  server.handle(BattleshipRelayService.new)
  puts "gRPC Battleship relay listening on port #{port}"
  server.run_till_terminated_or_interrupted([1, 'int', 'TERM'])
end
