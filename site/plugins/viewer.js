import Vue from 'vue'
import ViewerPlugin from 'v-viewer'
import 'viewerjs/dist/viewer.css'

Vue.use(ViewerPlugin, {
  defaultOptions: {
    zIndex: 9999,
    navbar: false,
    title: false,
    tooltip: false,
    movable: false,
    scalable: false,
    url: 'data-src',
  },
})
