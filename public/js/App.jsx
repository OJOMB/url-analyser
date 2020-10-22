import React, { Component } from 'react'
import ReactDOM from 'react-dom'
import Loader from 'react-loader-spinner'
import { parseJSON, checkStatus } from './helpers'

class App extends Component {
    render() {
        return (
            <div className="app">
                <NavBar />
                <UrlAnalyser />
            </div>
        );
    };
}

class UrlAnalyser extends Component {
    state = {
        loadingResult: false,
        showResult: false,
        result: null,
        url: null,
    };
    render() {
        if (this.state.showResult) {
            return (
                <ResultViewer 
                    result={this.state.result}
                    onCloseResult={this.handleCloseResults}
                />
            );
        } else if (this.state.loadingResult) {
            return (
                <LoadingScreen url={this.state.url}/>
            );
        } else {
            return (
                <UrlInputBox onSubmit={this.handleSubmit} />
            );
        }
    };
    handleSubmit = (e) => {
        this.setState({loadingResult: true, url: e.url})
        console.log("sending request to backend with data: " + e)
        return fetch(
            '/analyseUrl',
            {
                method: 'post',
                body: JSON.stringify({url: e}),
                headers: {
                    'Accept': 'application/json',
                    'Content-Type': 'application/json',
                },
            }
        ).then(checkStatus).then(parseJSON).then(this.handleResponse);
    };
    handleResponse = (res) => {
        console.log("received response from backend")
        console.log(res)
        this.setState(
            {
                loadingResult: false,
                showResult: true,
                result: res,
            }
        )
    };
    handleCloseResults = () => {
        this.setState(
            {
                loadingResult: false,
                showResult: false,
                result: null,
                url: null,
            }
        )
    }
}

class UrlInputBox extends Component {
    state = {
        userInput: ""
    };
    render() {
        return (
            <div className="url-input">
                <form>
                    <div className="align-middle">
                        <input
                            type="text"
                            className="form-control form-control-lg"
                            size={70}
                            placeholder="Enter URL"
                            onChange={(e) => this.setState({ userInput: e.target.value })}
                        />
                        <input className="submit-btn" type="submit" value="Analyse URL" onClick={() => this.props.onSubmit(this.state.userInput)}/> 
                    </div>
                </form>
            </div>
        );
    };
}

class NavBar extends Component {
    render() {
        return (
            <nav className="navbar">
                <div className="navbar-header">
                    <span className="navbar-brand">
                        <a href="/">URL Analyser</a>
                    </span>
                </div>
            </nav>
        );
    };
}

class LoadingScreen extends Component {
    render() {
        return (
            <div className="loader">
                <div className="loading-animation">
                    <Loader
                        className="loader"
                        type="Rings"
                        color="#e39734"
                        height={150}
                        width={150}
                    />
                </div>
                <p>Analysing URL: {this.props.url}</p>
                <p>This can take some time if the page contains lots of links...</p>
            </div>
        );
    };
}

class ResultViewer extends Component {
    render() {
        return (
            <div className="results-viewer">
                <h1>URL Analysis</h1>
                <p><b>URL:</b> <a href={this.props.result.url}>{this.props.result.url}</a></p>
                <p><b>Page Title:</b> {this.props.result.pageTitle}</p>
                <p><b>Links:</b></p>
                <ul>
                    <li>Total: {this.props.result.linksByType["Internal"]+this.props.result.linksByType["External"]}</li>
                    <li>Internal: {this.props.result.linksByType["Internal"]}</li>
                    <li>External: {this.props.result.linksByType["External"]}</li>
                    <li>Inaccessible: {this.props.result.inaccessibleLinks}</li>
                </ul>
                <p><b>Contains Login Form:</b> {this.props.result.loginForm ? "true": "false"}</p>
                <p><b>Headings:</b></p>
                <ul>
                    <li>h1: {this.props.result.headings["H1"]}</li>
                    <li>h2: {this.props.result.headings["H2"]}</li>
                    <li>h3: {this.props.result.headings["H3"]}</li>
                    <li>h4: {this.props.result.headings["H4"]}</li>
                    <li>h5: {this.props.result.headings["H5"]}</li>
                    <li>h6: {this.props.result.headings["H6"]}</li>

                </ul>
                <button className="btn btn-danger" onClick={this.props.onCloseResult}>x</button>
            </div>
        )
    }
}


ReactDOM.render(<App />, document.getElementById('root'));