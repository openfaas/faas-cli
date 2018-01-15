// Learn more about F# at http://fsharp.org

open System
open Function.Handler
open System.Text



let bufferInput () =
    let sb = StringBuilder()
    let rec buffer (s : string) =
        match s with
        | null -> sb.ToString()
        | s ->
            sb.Append(s) |> ignore
            buffer <| Console.ReadLine()
    buffer <| Console.ReadLine()

[<EntryPoint>]
let main argv =
    bufferInput ()
    |> handle
    0 // return an integer exit code
