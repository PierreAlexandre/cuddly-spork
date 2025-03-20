#!/usr/bin/env python3

import argparse
import asyncio
import logging
import random
import socket


async def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--port', '-p', type=int, default=8500, help="port to listen and make connections on")
    parser.add_argument('--num-connections', '-n', type=int, default=100, help="number of connections to open")
    parser.add_argument('--ipv', type=int, default=4, choices=[4, 6], help="IP version (4 or 6) to use")
    parser.add_argument('--verbose', '-v', action='store_true', help="enable verbose logging")
    args = parser.parse_args()

    level = logging.DEBUG if args.verbose else logging.INFO
    logging.basicConfig(format="%(message)s", level=level)

    if args.ipv == 4:
        host = "127.0.0.1"
        family = socket.AF_INET
    elif args.ipv == 6:
        host = "::1"
        family = socket.AF_INET6
    else:
        raise ValueError(args.ipv)

    loop = asyncio.get_running_loop()
    tasks = []

    started_event = asyncio.Event()
    server_task = asyncio.create_task(start_server(loop, host, args.port, family, started_event))
    tasks.append(server_task)

    await started_event.wait()
    logging.info(f"started server on {host}:{args.port}")

    for index in range(args.num_connections):
        client_task = asyncio.create_task(start_client(loop, index, host, args.port, family))
        tasks.append(client_task)

    logging.info(f"opened {args.num_connections} client connections")

    try:
        await asyncio.gather(*tasks)
    except asyncio.exceptions.CancelledError:
        logging.info("exiting")
        return


async def start_server(loop, host, port, family, started_event):
    server = await loop.create_server(
        lambda: ServerProtocol(),
        host=host, port=port, family=family)

    started_event.set()

    async with server:
        await server.serve_forever()


class ServerProtocol(asyncio.Protocol):
    def connection_made(self, transport):
        self.transport = transport
        peer = transport.get_extra_info("peername")
        logging.debug(f"server received connection: {peer}")

    def data_received(self, data):
        message = data.decode()
        logging.debug(f"server received message: {message}")
        self.transport.write(b"pong")


async def start_client(loop, client_idx, host, port, family):
    transport, protocol = await loop.create_connection(
        lambda: ClientProtocol(client_idx),
        host=host, port=port, family=family)

    message = f"ping {client_idx}".encode()
    try:
        while True:
            transport.write(message)
            delay = 10 + (random.random() * 10)
            await asyncio.sleep(delay)
    finally:
        transport.close()


class ClientProtocol(asyncio.Protocol):
    def __init__(self, client_idx):
        self.client_idx = client_idx

    def connection_made(self, transport):
        peer = transport.get_extra_info("peername")
        logging.debug(f"client {self.client_idx} made connection: {peer}")

    def connection_lost(self, exc):
        logging.debug(f"client {self.client_idx} closed connection: {exc}")


if __name__ == "__main__":
    asyncio.run(main())
