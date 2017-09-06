package Handler;

use JSON::XS;

sub handle {
    print encode_json( { 'echo' => shift } );
}

1;
