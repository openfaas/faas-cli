<?php

class Handler
{
    public function handle(string $data): void {
        echo json_encode([PHP_VERSION, $data]);
    }
}
