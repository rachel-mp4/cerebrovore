import { mount } from 'svelte'
import Chat from './Chat.svelte'

const eub = document.getElementById('eats-ur-brain')!
const ismoderator = eub.classList.contains("mod")
const app = mount(Chat, {
  target: eub,
  props: {
    ismoderator: ismoderator
  }
})

export default app
