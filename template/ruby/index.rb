require_relative 'function/handler'

req = ARGF.read

handler = Handler.new
res = handler.run req

puts res
