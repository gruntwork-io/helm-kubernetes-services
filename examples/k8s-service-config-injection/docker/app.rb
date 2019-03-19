# A sample backend app built on top of Ruby and Sinatra. It returns JSON.

require 'sinatra'
require 'json'

server_port = ENV['SERVER_PORT'] || 8080
server_text = ENV['SERVER_TEXT'] || 'Hello from backend'

set :port, server_port
set :bind, '0.0.0.0'

get '/' do
  content_type :json
  {:text => server_text}.to_json
end
