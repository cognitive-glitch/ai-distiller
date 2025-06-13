<?php

declare(strict_types=1);

interface Loggable
{
    public function log(string $message): void;
}

abstract class AbstractStorage
{
    protected string $storagePath;

    public function __construct(string $storagePath)
    {
        $this->storagePath = rtrim($storagePath, '/');
    }

    abstract protected function save(string $key, string $data): bool;
    
    final public function getStoragePath(): string
    {
        return $this->storagePath;
    }
}

class FileLogger extends AbstractStorage implements Loggable
{
    public function __construct(string $logDirectory)
    {
        parent::__construct($logDirectory);
    }

    public function log(string $message): void
    {
        $this->save('log_' . date('Y-m-d'), $message . PHP_EOL);
    }

    protected function save(string $key, string $data): bool
    {
        $file = $this->storagePath . '/' . $key . '.log';
        return file_put_contents($file, $data, FILE_APPEND) !== false;
    }
}