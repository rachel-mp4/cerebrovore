import { mount } from 'svelte'
import Watcher from './Watcher.svelte'

const bue = document.getElementById('brains-ur-eat')
const isforum = bue ? true : false
const app = mount(Watcher, {
  target: document.getElementById('watched-threads')!,
  props: {
    isforum: isforum
  }
})

export default app
