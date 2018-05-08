# L7 Load Balancing without a Service Mesh

This reposity contains the support code for my Kubecon Talk, [The Untapped
Power of Services - L7 Load Balancing Without a Service Mesh](
https://kccnceu18.sched.com/event/ENvv/the-untapped-power-of-services-l7-load-balancing-without-a-service-mesh-damien-lespiau-weaveworks-advanced-skill-level).

It implements cliend-side l7 load balancing of Kubernetes Services using
[consistent hashing with bounded loads](https://arxiv.org/abs/1608.01350).

This is prototype code but hopefully a decent start for anyone wanting to
implement similar load balancing for their own micro-services.

The talk is available in video:

<p align="center">
  <a href="http://www.youtube.com/watch?feature=player_embedded&v=PQnTBUr174M"
     target="_blank">
    <img src="http://img.youtube.com/vi/PQnTBUr174M/0.jpg" 
         alt="The talk in video" width="640" height="480" border="10" />
  </a>
</p>

