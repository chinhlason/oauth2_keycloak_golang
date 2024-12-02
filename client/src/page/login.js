import React from 'react';
import {useGoogleLogin} from "@react-oauth/google";
import request from "../utils/fetch";
import {useNavigate} from "react-router-dom";
import Cookies from "js-cookie";

const Login = () => {
    const nav = useNavigate();
    const handleGoogleLogin = useGoogleLogin({
        onSuccess: async (response) => {
            console.log('Google Login Success:', response);
            alert('success!');
            await request.post(`/callback?code=${response.code}`)
                .then((res) => {
                    const response = res.data;
                    console.log('response', response);
                    Cookies.set('token', response.access_token);
                    Cookies.set('refresh_data', response.refresh_token);
                    nav('/home');
                })
                .catch((error) => {
                    console.error('Google Login Failed:', error);
                    alert('fail!');
                });
        },
        onError: () => {
            console.error('Google Login Failed');
            alert('fail!');
        },
        flow: 'auth-code',
    });

    const handleFormSubmit = (e) => {

    };

    const sharedStyle = {
        width: '100%',
        padding: '10px',
        borderRadius: '5px',
        fontSize: '16px',
        boxSizing: 'border-box',
    };

    return (
        <div style={{ maxWidth: '400px', margin: '0 auto', marginTop: '100px', textAlign: 'center' }}>
            <h1>Login</h1>
            <form onSubmit={handleFormSubmit} style={{ marginBottom: '20px' }}>
                <div style={{ marginBottom: '15px' }}>
                    <input
                        type="email"
                        name="email"
                        placeholder="Email"
                        required
                        style={{
                            ...sharedStyle,
                            border: '1px solid #ccc',
                        }}
                    />
                </div>
                <div style={{ marginBottom: '15px' }}>
                    <input
                        type="password"
                        name="password"
                        placeholder="Password"
                        required
                        style={{
                            ...sharedStyle,
                            border: '1px solid #ccc',
                        }}
                    />
                </div>
                <button
                    type="submit"
                    style={{
                        ...sharedStyle,
                        backgroundColor: '#28a745',
                        color: '#fff',
                        border: 'none',
                        cursor: 'pointer',
                    }}
                >
                    Login
                </button>
            </form>

            <button
                onClick={handleGoogleLogin}
                style={{
                    ...sharedStyle,
                    backgroundColor: 'rgba(250,54,89,0.9)',
                    marginBottom: '20px',
                    color: '#fff',
                    border: 'none',
                    cursor: 'pointer',
                }}
            >
                Sign up
            </button>

            <button
                onClick={handleGoogleLogin}
                style={{
                    ...sharedStyle,
                    backgroundColor: '#4285F4',
                    color: '#fff',
                    border: 'none',
                    cursor: 'pointer',
                }}
            >
                Login with Google
            </button>
        </div>
    );
};

export default Login;
