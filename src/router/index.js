import { createRouter, createWebHistory } from "vue-router";
import MapView from "../views/MapView.vue"; // 地图页面
import AboutPage from "../views/AboutPage.vue"; // 示例其他页面

const routes = [
  {
    path: "/",
    name: "Map",
    component: MapView,
  },
  {
    path: "/about",
    name: "About",
    component: AboutPage,
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

export default router;
