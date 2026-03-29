from flask import Flask, render_template

app = Flask(__name__, template_folder=".")


def render_legacy_template(bgcolor: str, text: str, next_url: str, image: str):
    return render_template(
        "template.html",
        bgcolor=bgcolor,
        text=text,
        next_url=next_url,
        image=image,
    )


@app.get("/")
def main_page():
    return render_template("index.html")


@app.get("/ofcourse")
def ofcourse():
    return render_legacy_template("#FFFFFF", "#000000", "/", "ofcourse.jpg")


@app.get("/funnyman")
def funnyman():
    return render_legacy_template("#000000", "#FFFFFF", "brown", "funnyman.jpg")


@app.get("/brown")
def brown():
    return render_legacy_template("#000000", "#FFFFFF", "nurbs", "brown.jpg")


@app.get("/nurbs")
def nurbs():
    return render_legacy_template("#FFFFFF", "#000000", "thenextlevel", "nurbs.jpg")


@app.get("/thenextlevel")
def thenextlevel():
    return render_legacy_template("#FFFFFF", "#000000", "dog", "thenextlevel.jpg")


@app.get("/dog")
def dog():
    return render_legacy_template(
        "#000000",
        "#FFFFFF",
        "http://johnniemanzari.com",
        "dog.jpg",
    )


if __name__ == "__main__":
    app.run(host="127.0.0.1", port=8080, debug=True)
