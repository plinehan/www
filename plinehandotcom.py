import wsgiref.handlers
from google.appengine.ext import webapp
from google.appengine.ext.webapp import template

class MainPage(webapp.RequestHandler):
  def get(self):
    self.response.out.write(template.render('index.html', {}).decode())

class OfCourse(webapp.RequestHandler):
  def get(self):
    values = {
      'bgcolor':  '#FFFFFF',
      'text':     '#000000',
      'next_url': '/',
      'image':    'ofcourse.jpg'
      }
    self.response.out.write(template.render('template.html', values).decode())

class FunnyMan(webapp.RequestHandler):
  def get(self):
    values = {
      'bgcolor':  '#000000',
      'text':     '#FFFFFF',
      'next_url': 'brown',
      'image':    'funnyman.jpg'
      }
    self.response.out.write(template.render('template.html', values).decode())

class Brown(webapp.RequestHandler):
  def get(self):
    values = {
      'bgcolor':  '#000000',
      'text':     '#FFFFFF',
      'next_url': 'nurbs',
      'image':    'brown.jpg'
      }
    self.response.out.write(template.render('template.html', values).decode())

class Nurbs(webapp.RequestHandler):
  def get(self):
    values = {
      'bgcolor':  '#FFFFFF',
      'text':     '#000000',
      'next_url': 'thenextlevel',
      'image':    'nurbs.jpg'
      }
    self.response.out.write(template.render('template.html', values).decode())

class TheNextLevel(webapp.RequestHandler):
  def get(self):
    values = {
      'bgcolor':  '#FFFFFF',
      'text':     '#000000',
      'next_url': 'dog',
      'image':    'thenextlevel.jpg'
      }
    self.response.out.write(template.render('template.html', values).decode())

class Dog(webapp.RequestHandler):
  def get(self):
    values = {
      'bgcolor':  '#000000',
      'text':     '#FFFFFF',
      'next_url': 'http://www.johnniemanzari.com',
      'image':    'dog.jpg'
      }
    self.response.out.write(template.render('template.html', values).decode())
    
def main():
  application = webapp.WSGIApplication([('/', MainPage),
                                        ('/ofcourse', OfCourse),
                                        ('/funnyman', FunnyMan),
                                        ('/brown', Brown),
                                        ('/nurbs', Nurbs),
                                        ('/thenextlevel', TheNextLevel),
                                        ('/dog', Dog)],
                                       debug=True)
  wsgiref.handlers.CGIHandler().run(application)

if __name__ == '__main__':
  main()
