using UnityEngine;
using UnityEngine.Rendering;

namespace CyberWoodenFish.game
{
    [DisallowMultipleComponent]
    [RequireComponent(typeof(Rigidbody))]
    public sealed class TetheredTarget : MonoBehaviour
    {
        [Header("Tether")]
        [SerializeField] private Transform anchor;
        [SerializeField] private Rigidbody ropeBody;
        [SerializeField] private Vector3 localAttachmentPoint = new(0f, 0f, -0.7f);
        [SerializeField, Min(0.001f)] private float lineWidth = 0.015f;
        [SerializeField] private Color lineColor = Color.white;

        [Header("Hit")]
        [SerializeField, Min(0f)] private float hitImpulse = 30f;

        private Rigidbody _body;
        private LineRenderer _lineRenderer;
        private Material _lineMaterial;

        private void Awake()
        {
            _body = GetComponent<Rigidbody>();
            _body.interpolation = RigidbodyInterpolation.Interpolate;
            _body.collisionDetectionMode = CollisionDetectionMode.ContinuousDynamic;

            ConfigureRopePivot();
            ConfigureLineRenderer();
        }

        private void LateUpdate()
        {
            if (_lineRenderer == null || anchor == null) return;

            _lineRenderer.SetPosition(0, anchor.position);
            _lineRenderer.SetPosition(1, transform.TransformPoint(localAttachmentPoint));
        }

        public void ApplyHit(Vector3 hitPoint, Vector3 direction)
        {
            if (direction.sqrMagnitude <= Mathf.Epsilon) return;

            _body.WakeUp();
            _body.AddForceAtPosition(direction.normalized * hitImpulse, hitPoint, ForceMode.Impulse);
        }

        private void ConfigureRopePivot()
        {
            if (ropeBody == null) return;

            ropeBody.interpolation = RigidbodyInterpolation.Interpolate;

            var oldPivot = ropeBody.GetComponent<HingeJoint>();
            if (oldPivot == null) return;

            var worldPivot = ropeBody.transform.TransformPoint(oldPivot.anchor);
            var freePivot = ropeBody.GetComponent<ConfigurableJoint>();
            if (freePivot == null)
            {
                freePivot = ropeBody.gameObject.AddComponent<ConfigurableJoint>();
            }

            freePivot.connectedBody = oldPivot.connectedBody;
            freePivot.autoConfigureConnectedAnchor = false;
            freePivot.anchor = oldPivot.anchor;
            freePivot.connectedAnchor = oldPivot.connectedBody == null
                ? worldPivot
                : oldPivot.connectedBody.transform.InverseTransformPoint(worldPivot);
            freePivot.xMotion = ConfigurableJointMotion.Locked;
            freePivot.yMotion = ConfigurableJointMotion.Locked;
            freePivot.zMotion = ConfigurableJointMotion.Locked;
            freePivot.angularXMotion = ConfigurableJointMotion.Free;
            freePivot.angularYMotion = ConfigurableJointMotion.Free;
            freePivot.angularZMotion = ConfigurableJointMotion.Free;
            freePivot.projectionMode = JointProjectionMode.PositionAndRotation;
            freePivot.projectionDistance = 0.02f;
            freePivot.projectionAngle = 2f;
            freePivot.enableCollision = false;

            Destroy(oldPivot);
        }

        private void ConfigureLineRenderer()
        {
            _lineRenderer = GetComponent<LineRenderer>();
            if (_lineRenderer == null)
            {
                _lineRenderer = gameObject.AddComponent<LineRenderer>();
            }

            var shader = Shader.Find("Sprites/Default");
            if (shader == null)
            {
                shader = Shader.Find("Unlit/Color");
            }

            if (shader != null)
            {
                _lineMaterial = new Material(shader)
                {
                    color = lineColor,
                    hideFlags = HideFlags.HideAndDontSave
                };
                _lineRenderer.sharedMaterial = _lineMaterial;
            }

            _lineRenderer.positionCount = 2;
            _lineRenderer.useWorldSpace = true;
            _lineRenderer.startWidth = lineWidth;
            _lineRenderer.endWidth = lineWidth;
            _lineRenderer.startColor = lineColor;
            _lineRenderer.endColor = lineColor;
            _lineRenderer.numCapVertices = 4;
            _lineRenderer.alignment = LineAlignment.View;
            _lineRenderer.textureMode = LineTextureMode.Stretch;
            _lineRenderer.shadowCastingMode = ShadowCastingMode.Off;
            _lineRenderer.receiveShadows = false;
        }

        private void OnDestroy()
        {
            if (_lineMaterial != null)
            {
                Destroy(_lineMaterial);
            }
        }

        private void OnValidate()
        {
            lineWidth = Mathf.Max(0.001f, lineWidth);
            hitImpulse = Mathf.Max(0f, hitImpulse);
        }
    }
}
