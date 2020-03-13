# Qlik Sense installer documentation

## Local development of documentation

Documentation is built using [mkdocs](https://www.mkdocs.org/) and uses [Material for MKDocs theme](https://squidfunk.github.io/mkdocs-material/)

Requirements: Python and PIP or Docker

```console
pip install mkdocs
pip install mkdocs-material
```

View live changes locally at http://localhost:8000
```console
mkdocs serve
```

### Docker

```console
docker run --rm -it -p 8000:8000 -v ${PWD}:/docs squidfunk/mkdocs-material
```
