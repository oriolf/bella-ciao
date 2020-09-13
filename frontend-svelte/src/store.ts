import { writable } from 'svelte/store';
import type { User } from './types/models.type';

export const user = writable<User>(null);