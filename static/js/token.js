function isTokenExpired(token) {
    try {
        const payload = JSON.parse(atob(token.split('.')[1]));
        const currentTime = Math.floor(Date.now() / 1000);
        return payload.exp < currentTime;
    } catch (e) {
        console.error("Invalid token", e);
        return true;
    }
}

function checkToken() {
    const token = localStorage.getItem('token');
    const currentUrl = window.location.pathname ;
    
    if (token && isTokenExpired(token)) {
        localStorage.removeItem('token');
        console.log('Token is expired and has been removed');
        if(currentUrl !== '/') window.location.href = "/";
    } else if (!token) {
        console.log('No token found');
        if(currentUrl !== '/') window.location.href = "/";
    } else {
        console.log('Token is still valid');
        if(currentUrl === '/') window.location.href = "dashboard";;
    }
}

checkToken();
