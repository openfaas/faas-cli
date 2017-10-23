defmodule OpenFaaS.Function do
  def main(argv) do
    IO.puts "#{inspect argv}"
  end
end
