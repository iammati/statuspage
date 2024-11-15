#!/bin/sh

bun add npm -g
bun install --include=dev
bun update
bun run dev
