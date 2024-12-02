import Login from '../page/login';
import Home from '../page/home';

const routers = [
    { path : '/', component : <Login /> },
    { path : '/home', component : <Home />}
];

export default routers;

