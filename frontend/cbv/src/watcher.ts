import { mount } from 'svelte'
import Watcher from './Watcher.svelte'

const app = mount(Watcher, {
  target: document.getElementById('watched-threads')!,
})

export default app
