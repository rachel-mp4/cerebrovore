import { mount } from 'svelte'
import Chat from './Chat.svelte'

const eub = document.getElementById('eats-ur-brain')!
const ismoderator = eub.classList.contains("mod")
const nick = eub.getAttribute("data-nick")
const cs = eub.getAttribute("data-color")
const color = cs !== null ? parseInt(cs, 10) : null
const app = mount(Chat, {
  target: eub,
  props: {
    ismoderator: ismoderator,
    defaultnick: nick,
    defaultcolor: color
  }
})

export default app
