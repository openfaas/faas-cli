use lib qw(function local/lib/perl5 function/local/lib/perl5);
require Handler;

sub get_stdin {
    my $buf = "";
    while (<STDIN>) {
        $buf .= $_;
    }
    return $buf;
}

my $st = get_stdin();
Handler::handle($st);
