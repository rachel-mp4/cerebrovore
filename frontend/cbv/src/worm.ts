import { mount } from 'svelte'
import Worm from './Worm.svelte'

const app = mount(Worm, {
  target: document.getElementById('right-sidebar')!,
})

export default app
