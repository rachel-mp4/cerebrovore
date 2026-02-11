import { mount } from 'svelte'
import Chat from './Chat.svelte'

const app = mount(Chat, {
  target: document.getElementById('cbv')!,
})

export default app
