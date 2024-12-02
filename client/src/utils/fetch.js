import axios from 'axios';
import Cookies from 'js-cookie';

const request = axios.create({
    baseURL: 'http://localhost:2901',
    headers: {
        'Content-Type': 'application/json',
    },
});

request.interceptors.request.use(
    (config) => {
        const token = Cookies.get('token');
        if (token) {
            config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
    },
    (error) => {
        return Promise.reject(error);
    }
);

export const get = async (path, options =  {}) => {
    const response = await request.get(path, options);
    return response.data;
}

export const post = async (path, data, options = {}) => {
    const response = await request.post(path, data, options);
    return response.data;
}

export default request;