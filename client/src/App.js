import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import routers from './router/router';

function App() {
    return (
        <Router>
            <div className="App">
                <Routes>
                    {
                        routers.map((router, index) => {
                            return (
                                <Route
                                    key={index}
                                    path={router.path}
                                    element={router.component}
                                />
                            )
                        })
                    }
                </Routes>
            </div>
        </Router>
    );
}

export default App;
