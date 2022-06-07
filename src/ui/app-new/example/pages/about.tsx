import * as React from 'react';
import {Link} from "react-router-dom";

const AboutPage: React.FC = () => {
    return (
        <>
            <main>
                <h2>Who are we?</h2>
                <p>
                    That feels like an existential question, don't you
                    think?
                </p>
            </main>
            <nav>
                <Link to="/">Home</Link>
            </nav>
        </>
    );
}

export default AboutPage;
