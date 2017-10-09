defmodule OpenFaaS.Function.Mixfile do
  use Mix.Project

  def project do
    [ app: :openfaas_function,
      version: "0.0.0",
      elixir: "~> 1.5",
      build_embedded: Mix.env == :prod,
      start_permanent: Mix.env == :prod,
      deps: deps(),
      escript: escript() ]
  end

  def escript do
    [ main_module: OpenFaaS.Function ]
  end

  def deps do
    [ ]
  end
end
