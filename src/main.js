import { createApp } from 'vue';
import App from './App.vue';
import router from './router';
import ElementPlus from 'element-plus';
import 'element-plus/dist/index.css';
import Echarts from "vue-echarts";
import * as echarts from "echarts";

const app = createApp(App);

app.use(router);
app.use(ElementPlus);
app.component("v-chart", Echarts);
app.config.globalProperties.$echarts = echarts;

app.mount('#app');