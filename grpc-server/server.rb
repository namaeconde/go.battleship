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
    game_id = @mutex.synchronize do
      id = generate_game_id
      @sessions[id] = { player_b_joined: false, streams: [] }
      id
    end
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

  def game_stream(requests, _call)
    queue   = Queue.new
    game_id = nil

    Thread.new do
      begin
        requests.each do |msg|
          if game_id.nil?
            game_id = msg.game_id
            break unless @mutex.synchronize { @sessions.key?(game_id) }
            unless register_stream(game_id, queue)
              break  # ensure will fire :eof
            end
          end
          relay_to_opponent(game_id, queue, msg)
        end
      ensure
        deregister_stream(game_id, queue) if game_id
        queue << :eof
      end
    end

    Enumerator.new do |y|
      loop do
        msg = queue.pop
        break if msg == :eof
        y << msg
      end
    end
  end

  private

  def generate_game_id
    chars = ('A'..'Z').to_a + ('0'..'9').to_a
    loop do
      id = 6.times.map { chars.sample }.join
      return id unless @sessions.key?(id)
    end
  end

  def register_stream(game_id, queue)
    @mutex.synchronize do
      session = @sessions[game_id]
      return false unless session
      return false if session[:streams].size >= 2
      session[:streams] << queue
      true
    end
  end

  def deregister_stream(game_id, queue)
    @mutex.synchronize do
      session = @sessions[game_id]
      return unless session
      session[:streams].delete(queue)
      @sessions.delete(game_id) if session[:streams].empty?
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
