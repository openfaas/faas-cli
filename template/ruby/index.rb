# Copyright (c) Alex Ellis 2017. All rights reserved.
# Licensed under the MIT license. See LICENSE file in the project root for full license information.

require_relative 'function/handler'

req = ARGF.read

handler = Handler.new
res = handler.run req

puts res
